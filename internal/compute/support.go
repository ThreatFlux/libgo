package compute

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ResourceTracker tracks resource usage across all instances
type ResourceTracker struct {
	mu        sync.RWMutex
	instances map[string]*ComputeInstance
}

// NewResourceTracker creates a new resource tracker
func NewResourceTracker() *ResourceTracker {
	return &ResourceTracker{
		instances: make(map[string]*ComputeInstance),
	}
}

// AddInstance adds an instance to tracking
func (rt *ResourceTracker) AddInstance(instance *ComputeInstance) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.instances[instance.ID] = instance
}

// UpdateInstance updates an instance in tracking
func (rt *ResourceTracker) UpdateInstance(instance *ComputeInstance) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.instances[instance.ID] = instance
}

// RemoveInstance removes an instance from tracking
func (rt *ResourceTracker) RemoveInstance(id string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	delete(rt.instances, id)
}

// GetTotalResources returns total allocated resources
func (rt *ResourceTracker) GetTotalResources() ComputeResources {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	var total ComputeResources
	for _, instance := range rt.instances {
		if instance.State == StateRunning {
			total.CPU.Cores += instance.Resources.CPU.Cores
			total.Memory.Limit += instance.Resources.Memory.Limit
		}
	}

	return total
}

// QuotaManager manages resource quotas for users
type QuotaManager struct {
	mu     sync.RWMutex
	quotas map[uint]*ResourceQuotas
}

// NewQuotaManager creates a new quota manager
func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		quotas: make(map[uint]*ResourceQuotas),
	}
}

// SetQuotas sets quotas for a user
func (qm *QuotaManager) SetQuotas(userID uint, quotas ResourceQuotas) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.quotas[userID] = &quotas
	return nil
}

// GetQuotas gets quotas for a user
func (qm *QuotaManager) GetQuotas(userID uint) *ResourceQuotas {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	if quota, exists := qm.quotas[userID]; exists {
		return quota
	}

	// Return default quotas if none set
	return &ResourceQuotas{
		UserID:          userID,
		MaxInstances:    10,
		MaxCPUCores:     8.0,
		MaxMemoryGB:     32,
		MaxStorageGB:    500,
		MaxNetworks:     10,
		AllowedBackends: []ComputeBackend{BackendKVM, BackendDocker},
		AllowedTypes:    []ComputeInstanceType{InstanceTypeVM, InstanceTypeContainer},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// CheckQuota checks if a request would exceed quotas
func (qm *QuotaManager) CheckQuota(ctx context.Context, req ComputeInstanceRequest) error {
	quota := qm.GetQuotas(req.UserID)

	// Check backend allowed
	backendAllowed := false
	for _, backend := range quota.AllowedBackends {
		if backend == req.Backend {
			backendAllowed = true
			break
		}
	}
	if !backendAllowed {
		return fmt.Errorf("backend %s not allowed for user", req.Backend)
	}

	// Check instance type allowed
	typeAllowed := false
	for _, instanceType := range quota.AllowedTypes {
		if instanceType == req.Type {
			typeAllowed = true
			break
		}
	}
	if !typeAllowed {
		return fmt.Errorf("instance type %s not allowed for user", req.Type)
	}

	// TODO: Check actual resource usage against quotas
	// This would require querying current instances for the user

	return nil
}

// EventBus handles instance events
type EventBus struct {
	mu          sync.RWMutex
	events      []InstanceEvent
	subscribers map[string][]chan InstanceEvent
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		events:      make([]InstanceEvent, 0),
		subscribers: make(map[string][]chan InstanceEvent),
	}
}

// Emit emits an event
func (eb *EventBus) Emit(event InstanceEvent) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Store event
	eb.events = append(eb.events, event)

	// Keep only last 1000 events
	if len(eb.events) > 1000 {
		eb.events = eb.events[len(eb.events)-1000:]
	}

	// Notify subscribers
	if subscribers, exists := eb.subscribers[event.InstanceID]; exists {
		for _, ch := range subscribers {
			select {
			case ch <- event:
			default:
				// Channel full, skip
			}
		}
	}

	// Notify global subscribers
	if subscribers, exists := eb.subscribers["*"]; exists {
		for _, ch := range subscribers {
			select {
			case ch <- event:
			default:
				// Channel full, skip
			}
		}
	}
}

// GetEvents gets historical events for an instance
func (eb *EventBus) GetEvents(instanceID string, opts EventOptions) []*InstanceEvent {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	var filtered []*InstanceEvent
	for i := len(eb.events) - 1; i >= 0; i-- {
		event := eb.events[i]

		// Filter by instance ID
		if instanceID != "*" && event.InstanceID != instanceID {
			continue
		}

		// Filter by time range
		if opts.Since != nil && event.Timestamp.Before(opts.Since.Time) {
			continue
		}
		if opts.Until != nil && event.Timestamp.After(opts.Until.Time) {
			continue
		}

		// Filter by event types
		if len(opts.Types) > 0 {
			typeMatch := false
			for _, t := range opts.Types {
				if event.Type == t {
					typeMatch = true
					break
				}
			}
			if !typeMatch {
				continue
			}
		}

		eventCopy := event
		filtered = append(filtered, &eventCopy)

		// Apply limit
		if opts.Limit > 0 && len(filtered) >= opts.Limit {
			break
		}
	}

	return filtered
}

// StreamEvents streams events for an instance
func (eb *EventBus) StreamEvents(instanceID string, opts EventOptions) <-chan InstanceEvent {
	ch := make(chan InstanceEvent, 100) // Buffer events

	eb.mu.Lock()
	if eb.subscribers[instanceID] == nil {
		eb.subscribers[instanceID] = make([]chan InstanceEvent, 0)
	}
	eb.subscribers[instanceID] = append(eb.subscribers[instanceID], ch)
	eb.mu.Unlock()

	// Send historical events if requested
	if !opts.Follow {
		go func() {
			events := eb.GetEvents(instanceID, opts)
			for _, event := range events {
				select {
				case ch <- *event:
				default:
					// Channel full
				}
			}
			close(ch)
		}()
	}

	return ch
}
