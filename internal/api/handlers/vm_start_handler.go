package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/pkg/logger"
)

// StartVMResponse represents the response for a VM start request.
type StartVMResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// StartVM handles requests to start a VM.
func (h *VMHandler) StartVM(c *gin.Context) {
	// Get VM name from URL path.
	name := c.Param("name")
	if name == "" {
		HandleError(c, ErrInvalidInput)
		return
	}

	// Get context logger.
	contextLogger := getContextLogger(c, h.logger)
	contextLogger = contextLogger.WithFields(logger.String("vmName", name))

	// Start VM.
	err := h.vmManager.Start(c.Request.Context(), name)
	if err != nil {
		contextLogger.Error("Failed to start VM",
			logger.Error(err))
		HandleError(c, err)
		return
	}

	// Log success.
	contextLogger.Info("VM started successfully")

	// Return response.
	c.JSON(http.StatusOK, StartVMResponse{
		Success: true,
		Message: "VM started successfully",
	})
}
