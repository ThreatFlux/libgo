package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/pkg/logger"
)

// DeleteVMResponse represents the response for a VM deletion request
type DeleteVMResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// DeleteVM handles requests to delete a VM
func (h *VMHandler) DeleteVM(c *gin.Context) {
	// Get VM name from URL path
	name := c.Param("name")
	if name == "" {
		HandleError(c, ErrInvalidInput)
		return
	}

	// Get context logger
	contextLogger := getContextLogger(c, h.logger)
	contextLogger = contextLogger.WithFields(logger.String("vmName", name))

	// Check if force delete is requested
	force := false
	if forceStr := c.Query("force"); forceStr == "true" {
		force = true
	}

	// Delete VM
	// Force parameter is currently not used in the interface
	// but we keep it in the handler for future implementation
	err := h.vmManager.Delete(c.Request.Context(), name)

	if err != nil {
		contextLogger.Error("Failed to delete VM",
			logger.Bool("force", force),
			logger.Error(err))
		HandleError(c, err)
		return
	}

	// Log success
	contextLogger.Info("VM deleted successfully",
		logger.Bool("force", force))

	// Return response
	c.JSON(http.StatusOK, DeleteVMResponse{
		Success: true,
		Message: "VM deleted successfully",
	})
}
