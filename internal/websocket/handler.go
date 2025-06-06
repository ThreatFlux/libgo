package websocket

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	vmmodels "github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/pkg/logger"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 8192
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development
		// Note: In production, this should be more restrictive to verify origin
		return true
	},
}

// Handler represents a WebSocket handler.
type Handler struct {
	hub    *Hub
	logger logger.Logger
}

// NewHandler creates a new WebSocket handler.
func NewHandler(logger logger.Logger) *Handler {
	hub := NewHub()
	go hub.Run()

	return &Handler{
		hub:    hub,
		logger: logger,
	}
}

// HandleVM handles WebSocket connections for VM monitoring.
func (h *Handler) HandleVM(c *gin.Context) {
	h.handleConnection(c, false)
}

// HandleVMConsole handles WebSocket connections for VM console.
func (h *Handler) HandleVMConsole(c *gin.Context) {
	h.handleConnection(c, true)
}

// handleConnection handles the WebSocket connection.
func (h *Handler) handleConnection(c *gin.Context, isConsole bool) {
	// Get VM name from path
	vmName := c.Param("name")
	if vmName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VM name is required"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		userIDStr = "unknown"
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection to WebSocket",
			logger.String("vmName", vmName),
			logger.String("userID", userIDStr),
			logger.Error(err))
		return
	}

	// Create client
	client := &Client{
		Conn:       conn,
		Send:       make(chan *Message, 256),
		UserID:     userIDStr,
		VMName:     vmName,
		IsConsole:  isConsole,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	// Register client with hub
	h.hub.register <- client

	// Start goroutines for reading and writing
	go h.readPump(client)
	go h.writePump(client)

	// Log connection
	h.logger.Info("WebSocket connection established",
		logger.String("vmName", vmName),
		logger.String("userID", userIDStr),
		logger.Bool("isConsole", isConsole))

	// Send welcome message
	client.Send <- NewMessage(MessageTypeConnection, map[string]interface{}{
		"status":  "connected",
		"vmName":  vmName,
		"message": fmt.Sprintf("Connected to %s WebSocket", vmName),
	})
}

// readPump pumps messages from the WebSocket connection to the hub.
func (h *Handler) readPump(client *Client) {
	defer func() {
		h.hub.unregister <- client
		client.Conn.Close()
		h.logger.Info("WebSocket connection closed",
			logger.String("vmName", client.VMName),
			logger.String("userID", client.UserID))
	}()

	client.Conn.SetReadLimit(maxMessageSize)
	if err := client.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		h.logger.Debug("Failed to set read deadline", logger.Error(err))
	}
	client.Conn.SetPongHandler(func(string) error {
		if err := client.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			h.logger.Debug("Failed to set read deadline in pong handler", logger.Error(err))
		}
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket read error",
					logger.String("vmName", client.VMName),
					logger.String("userID", client.UserID),
					logger.Error(err))
			}
			break
		}

		// Update last active time
		client.LastActive = time.Now()

		// Parse message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			h.logger.Error("Failed to parse WebSocket message",
				logger.String("vmName", client.VMName),
				logger.String("userID", client.UserID),
				logger.Error(err))

			// Send error message to client
			client.Send <- ErrorMessage("INVALID_MESSAGE", "Failed to parse message")
			continue
		}

		// Handle message based on type
		switch msg.Type {
		case MessageTypeCommand:
			h.handleCommand(client, &msg)
		case MessageTypeConsoleIn:
			if client.IsConsole {
				h.handleConsoleInput(client, &msg)
			} else {
				client.Send <- ErrorMessage("INVALID_MESSAGE_TYPE", "Console input not allowed on this connection")
			}
		case MessageTypeHeartbeat:
			// Just acknowledge heartbeat
			client.Send <- NewMessage(MessageTypeHeartbeat, map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
			})
		default:
			h.logger.Warn("Unknown WebSocket message type",
				logger.String("vmName", client.VMName),
				logger.String("userID", client.UserID),
				logger.String("messageType", string(msg.Type)))

			client.Send <- ErrorMessage("UNKNOWN_MESSAGE_TYPE", "Unknown message type")
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
func (h *Handler) writePump(client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !h.handleMessage(client, message, ok) {
				return
			}
		case <-ticker.C:
			if !h.handlePing(client) {
				return
			}
		}
	}
}

// handleMessage processes a message from the client channel.
func (h *Handler) handleMessage(client *Client, message *Message, ok bool) bool {
	if err := client.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		h.logger.Debug("Failed to set write deadline", logger.Error(err))
	}

	if !ok {
		return h.closeConnection(client)
	}

	// Marshal and write the primary message
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal WebSocket message",
			logger.String("vmName", client.VMName),
			logger.String("userID", client.UserID),
			logger.Error(err))
		return false
	}

	w, err := client.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return false
	}

	if _, err := w.Write(data); err != nil {
		h.logger.Debug("Failed to write message data", logger.Error(err))
		return false
	}

	// Write any queued messages
	if !h.writeQueuedMessages(client, w) {
		return false
	}

	if err := w.Close(); err != nil {
		return false
	}

	return true
}

// closeConnection handles closing the WebSocket connection gracefully.
func (h *Handler) closeConnection(client *Client) bool {
	if err := client.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
		h.logger.Debug("Failed to write close message", logger.Error(err))
	}
	return false
}

// writeQueuedMessages writes any additional queued messages to the current writer.
func (h *Handler) writeQueuedMessages(client *Client, w io.WriteCloser) bool {
	n := len(client.Send)
	for i := 0; i < n; i++ {
		queuedMessage := <-client.Send
		data, err := json.Marshal(queuedMessage)
		if err != nil {
			h.logger.Error("Failed to marshal WebSocket message",
				logger.String("vmName", client.VMName),
				logger.String("userID", client.UserID),
				logger.Error(err))
			continue
		}

		if _, err := w.Write([]byte("\n")); err != nil {
			h.logger.Debug("Failed to write newline", logger.Error(err))
			continue
		}

		if _, err := w.Write(data); err != nil {
			h.logger.Debug("Failed to write queued message data", logger.Error(err))
			continue
		}
	}
	return true
}

// handlePing sends a ping message to keep the connection alive.
func (h *Handler) handlePing(client *Client) bool {
	if err := client.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		h.logger.Debug("Failed to set write deadline for ping", logger.Error(err))
	}

	if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		return false
	}

	return true
}

// handleCommand handles VM commands.
func (h *Handler) handleCommand(client *Client, msg *Message) {
	// Extract command information
	action, ok := msg.Data["action"].(string)
	if !ok {
		client.Send <- ErrorMessage("INVALID_COMMAND", "Missing or invalid action")
		return
	}

	requestID, ok := msg.Data["requestId"].(string)
	if !ok {
		requestID = ""
	}
	if requestID == "" {
		requestID = fmt.Sprintf("cmd-%d", time.Now().UnixNano())
	}

	h.logger.Info("Received VM command",
		logger.String("vmName", client.VMName),
		logger.String("userID", client.UserID),
		logger.String("action", action),
		logger.String("requestId", requestID))

	// Note: Command handling implementation needed for VM operations
	// For now, just acknowledge the command
	client.Send <- ResponseMessage(requestID, true, fmt.Sprintf("Command '%s' acknowledged", action))
}

// handleConsoleInput handles VM console input.
func (h *Handler) handleConsoleInput(client *Client, msg *Message) {
	content, ok := msg.Data["content"].(string)
	if !ok {
		client.Send <- ErrorMessage("INVALID_CONSOLE_INPUT", "Missing or invalid content")
		return
	}

	h.logger.Debug("Received console input",
		logger.String("vmName", client.VMName),
		logger.String("userID", client.UserID),
		logger.Int("contentLength", len(content)))

	// Note: Console input handling implementation needed for VM console access
	// For now, just echo the input back
	client.Send <- ConsoleMessage(content, false)
}

// SendVMStatus sends a VM status update to all clients connected to the VM.
func (h *Handler) SendVMStatus(vmName string, status vmmodels.VMStatus, lastChange time.Time, uptime int64) {
	msg := StatusMessage(status, lastChange, uptime)
	h.hub.SendToVM(vmName, msg)
}

// SendVMMetrics sends VM metrics to all clients connected to the VM.
func (h *Handler) SendVMMetrics(vmName string, cpu float64, memory, memoryTotal, rxBytes, txBytes, readBytes, writeBytes uint64) {
	msg := MetricsMessage(cpu, memory, memoryTotal, rxBytes, txBytes, readBytes, writeBytes)
	h.hub.SendToVM(vmName, msg)
}

// SendVMConsoleOutput sends VM console output to console clients.
func (h *Handler) SendVMConsoleOutput(vmName string, content string, eof bool) {
	msg := ConsoleMessage(content, eof)

	// Only send to console clients
	clients := h.hub.vmClients[vmName]
	for _, client := range clients {
		if client.IsConsole {
			select {
			case client.Send <- msg:
			default:
				close(client.Send)
				delete(h.hub.clients, client)
			}
		}
	}
}
