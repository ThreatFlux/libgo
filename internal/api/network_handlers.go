package api

import "github.com/gin-gonic/gin"

// Handler is a common interface for all handlers
type Handler interface {
	Handle(c *gin.Context)
}

// NetworkHandlers holds all network-related handlers
type NetworkHandlers struct {
	List   Handler
	Create Handler
	Get    Handler
	Update Handler
	Delete Handler
	Start  Handler
	Stop   Handler
}
