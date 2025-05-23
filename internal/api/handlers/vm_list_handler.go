package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	vmmodels "github.com/threatflux/libgo/internal/models/vm"
	vmservice "github.com/threatflux/libgo/internal/vm"
	"github.com/threatflux/libgo/pkg/logger"
)

// VMHandler handles VM-related requests
type VMHandler struct {
	vmManager vmservice.Manager
	logger    logger.Logger
}

// NewVMHandler creates a new VMHandler
func NewVMHandler(vmManager vmservice.Manager, logger logger.Logger) *VMHandler {
	return &VMHandler{
		vmManager: vmManager,
		logger:    logger,
	}
}

// GetVMManager returns the VM manager instance
func (h *VMHandler) GetVMManager() vmservice.Manager {
	return h.vmManager
}

// ListVMsResponse represents the response for a VM listing request
type ListVMsResponse struct {
	VMs      interface{} `json:"vms"`
	Count    int         `json:"count"`
	PageSize int         `json:"pageSize,omitempty"`
	Page     int         `json:"page,omitempty"`
}

// ListVMs handles requests to list all VMs
func (h *VMHandler) ListVMs(c *gin.Context) {
	// Get page and page size parameters for pagination
	page := 1
	pageSize := 50

	if pageStr := c.Query("page"); pageStr != "" {
		if parsedPage, err := parseInt(pageStr, 1, 1000); err == nil {
			page = parsedPage
		} else {
			HandleError(c, ErrInvalidInput)
			return
		}
	}

	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if parsedPageSize, err := parseInt(pageSizeStr, 1, 100); err == nil {
			pageSize = parsedPageSize
		} else {
			HandleError(c, ErrInvalidInput)
			return
		}
	}

	// Get context logger if available
	contextLogger := getContextLogger(c, h.logger)

	// Get VMs from manager
	// Note: Filter functionality would be implemented here in the future
	vms, err := h.vmManager.List(c.Request.Context())
	if err != nil {
		contextLogger.Error("Failed to list VMs",
			logger.Error(err))
		HandleError(c, err)
		return
	}

	// Paginate results (simple implementation)
	// In a real app, pagination would likely be done at the database level
	totalCount := len(vms)
	start := (page - 1) * pageSize
	end := start + pageSize

	// Ensure end doesn't exceed the total count
	if end > totalCount {
		end = totalCount
	}

	// Ensure start is valid
	if start >= totalCount {
		// Return empty array if page is beyond available data
		vms = []*vmmodels.VM{}
	} else {
		// Extract the slice for this page
		paginatedVMs := vms[start:end]
		vms = paginatedVMs
	}

	// Create response
	response := ListVMsResponse{
		VMs:      vms,
		Count:    totalCount,
		PageSize: pageSize,
		Page:     page,
	}

	contextLogger.Info("Listed VMs successfully",
		logger.Int("count", totalCount),
		logger.Int("page", page),
		logger.Int("pageSize", pageSize))

	c.JSON(http.StatusOK, response)
}

// Helper functions

// getVMFilterFromQuery extracts VM filter parameters from the query string
func getVMFilterFromQuery(c *gin.Context) map[string]string {
	filter := make(map[string]string)

	// Add supported filter parameters
	if name := c.Query("name"); name != "" {
		filter["name"] = name
	}

	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	if tags := c.Query("tags"); tags != "" {
		filter["tags"] = tags
	}

	return filter
}

// parseInt parses a string to an integer with min/max bounds
func parseInt(value string, min, max int) (int, error) {
	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return 0, ErrInvalidInput
	}

	if result < min {
		return min, nil
	}

	if result > max {
		return max, nil
	}

	return result, nil
}

// getContextLogger gets the logger from the context or falls back to the provided logger
func getContextLogger(c *gin.Context, defaultLogger logger.Logger) logger.Logger {
	if loggerInstance, exists := c.Get("logger"); exists {
		if contextLogger, ok := loggerInstance.(logger.Logger); ok {
			return contextLogger
		}
	}
	return defaultLogger
}
