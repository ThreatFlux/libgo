package api

import (
	"github.com/threatflux/libgo/internal/api/handlers"
)

// OVSHandlers holds all OVS-related handlers
type OVSHandlers struct {
	// Bridge handlers
	CreateBridge *handlers.OVSBridgeCreateHandler
	ListBridges  *handlers.OVSBridgeListHandler
	GetBridge    *handlers.OVSBridgeGetHandler
	DeleteBridge *handlers.OVSBridgeDeleteHandler

	// Port handlers
	CreatePort *handlers.OVSPortCreateHandler
	ListPorts  *handlers.OVSPortListHandler
	DeletePort *handlers.OVSPortDeleteHandler

	// Flow handlers
	CreateFlow *handlers.OVSFlowCreateHandler
}
