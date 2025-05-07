package websocket

import (
	"time"

	"github.com/gorilla/websocket"
	vmmodels "github.com/threatflux/libgo/internal/models/vm"
)

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// Message types
	MessageTypeStatus     MessageType = "status"
	MessageTypeMetrics    MessageType = "metrics"
	MessageTypeCommand    MessageType = "command"
	MessageTypeResponse   MessageType = "response"
	MessageTypeConsole    MessageType = "console"
	MessageTypeConsoleIn  MessageType = "console_input"
	MessageTypeError      MessageType = "error"
	MessageTypeHeartbeat  MessageType = "heartbeat"
	MessageTypeConnection MessageType = "connection"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType          `json:"type"`
	Timestamp time.Time            `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewMessage creates a new message with the current timestamp
func NewMessage(msgType MessageType, data map[string]interface{}) *Message {
	return &Message{
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// StatusMessage creates a new status message
func StatusMessage(status vmmodels.VMStatus, lastChange time.Time, uptime int64) *Message {
	return NewMessage(MessageTypeStatus, map[string]interface{}{
		"status":         status,
		"lastStateChange": lastChange.Format(time.RFC3339),
		"uptime":         uptime,
	})
}

// MetricsMessage creates a new metrics message
func MetricsMessage(cpu float64, memory uint64, memoryTotal uint64, 
	rxBytes, txBytes, readBytes, writeBytes uint64) *Message {
	return NewMessage(MessageTypeMetrics, map[string]interface{}{
		"cpu": map[string]interface{}{
			"utilization": cpu,
		},
		"memory": map[string]interface{}{
			"used":  memory,
			"total": memoryTotal,
		},
		"network": map[string]interface{}{
			"rxBytes": rxBytes,
			"txBytes": txBytes,
		},
		"disk": map[string]interface{}{
			"readBytes":  readBytes,
			"writeBytes": writeBytes,
		},
	})
}

// ErrorMessage creates a new error message
func ErrorMessage(code string, message string) *Message {
	return NewMessage(MessageTypeError, map[string]interface{}{
		"code":    code,
		"message": message,
	})
}

// ConsoleMessage creates a new console message
func ConsoleMessage(content string, eof bool) *Message {
	return NewMessage(MessageTypeConsole, map[string]interface{}{
		"content": content,
		"eof":     eof,
	})
}

// ResponseMessage creates a new response message
func ResponseMessage(requestID string, success bool, message string) *Message {
	return NewMessage(MessageTypeResponse, map[string]interface{}{
		"requestId": requestID,
		"success":   success,
		"message":   message,
	})
}

// Client represents a WebSocket client connection
type Client struct {
	Conn       *websocket.Conn
	Send       chan *Message
	UserID     string
	VMName     string
	IsConsole  bool
	CreatedAt  time.Time
	LastActive time.Time
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan *Message

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// VM name to clients mapping for targeted messages
	vmClients map[string][]*Client
}

// NewHub creates a new hub instance
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		vmClients:  make(map[string][]*Client),
	}
}

// Run starts the hub to handle client connections and messages
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			// Add to VM specific clients
			h.vmClients[client.VMName] = append(h.vmClients[client.VMName], client)
		
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				
				// Remove from VM specific clients
				clients := h.vmClients[client.VMName]
				for i, c := range clients {
					if c == client {
						h.vmClients[client.VMName] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
				
				// Clean up empty VM entries
				if len(h.vmClients[client.VMName]) == 0 {
					delete(h.vmClients, client.VMName)
				}
			}
		
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// SendToVM sends a message to all clients connected to a specific VM
func (h *Hub) SendToVM(vmName string, message *Message) {
	clients := h.vmClients[vmName]
	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.clients, client)
		}
	}
}