package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/pkg/logger"
)

// StopVMResponse represents the response for a VM stop request
type StopVMResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// StopVM handles requests to stop a VM
func (h *VMHandler) StopVM(c *gin.Context) {
	// Get VM name from URL path
	name := c.Param("name")
	if name == "" {
		HandleError(c, ErrInvalidInput)
		return
	}

	// Get context logger
	contextLogger := getContextLogger(c, h.logger)
	contextLogger = contextLogger.WithFields(logger.String("vmName", name))

	// Check if force stop is requested
	force := false
	if forceStr := c.Query("force"); forceStr == "true" {
		force = true
	}

	// Get timeout parameter (default: 30 seconds)
	timeout := 30
	if timeoutStr := c.Query("timeout"); timeoutStr != "" {
		if parsedTimeout, err := parseInt(timeoutStr, 0, 300); err == nil {
			timeout = parsedTimeout
		} else {
			HandleError(c, ErrInvalidInput)
			return
		}
	}

	// Stop VM
	// Note: Force option and timeout are not currently implemented in the interface
	// but we keep it in the handler for future implementation
	err := h.vmManager.Stop(c.Request.Context(), name)

	if err != nil {
		contextLogger.Error("Failed to stop VM",
			logger.Bool("force", force),
			logger.Int("timeout", timeout),
			logger.Error(err))
		HandleError(c, err)
		return
	}

	// Log success
	contextLogger.Info("VM stopped successfully",
		logger.Bool("force", force),
		logger.Int("timeout", timeout))

	// Return response
	c.JSON(http.StatusOK, StopVMResponse{
		Success: true,
		Message: "VM stopped successfully",
	})
}
