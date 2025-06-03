package websocket

import (
	"context"
	"sync"
	"time"

	vmmodels "github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/pkg/logger"
)

// VMMonitorInterval is the interval between VM metric collections.
const VMMonitorInterval = 5 * time.Second

// VMMonitor monitors VM metrics and broadcasts updates.
type VMMonitor struct {
	monitoredVMs     map[string]*monitoredVM
	handler          *Handler
	logger           logger.Logger
	vmManager        VMManager
	shutdown         chan struct{}
	monitoredVMsLock sync.RWMutex
}

// monitoredVM holds monitoring state for a VM.
type monitoredVM struct {
	cancelContext context.CancelFunc
	lastChecked   time.Time
	stateChanged  time.Time
	startTime     time.Time
	name          string
	lastStatus    vmmodels.VMStatus
	clientCount   int
	isMonitoring  bool
}

// VMManager is the interface for VM operations.
type VMManager interface {
	Get(ctx context.Context, name string) (*vmmodels.VM, error)
	GetMetrics(ctx context.Context, name string) (*VMMetrics, error)
}

// VMMetrics contains VM metrics data.
type VMMetrics struct {
	CPU struct {
		Utilization float64
	}
	Memory struct {
		Used  uint64
		Total uint64
	}
	Network struct {
		RxBytes uint64
		TxBytes uint64
	}
	Disk struct {
		ReadBytes  uint64
		WriteBytes uint64
	}
}

// NewVMMonitor creates a new VM monitor.
func NewVMMonitor(handler *Handler, vmManager VMManager, logger logger.Logger) *VMMonitor {
	return &VMMonitor{
		handler:      handler,
		logger:       logger,
		vmManager:    vmManager,
		monitoredVMs: make(map[string]*monitoredVM),
		shutdown:     make(chan struct{}),
	}
}

// Start starts the VM monitor.
func (m *VMMonitor) Start() {
	m.logger.Info("Starting VM monitor")
	go m.cleanupRoutine()
}

// Stop stops the VM monitor.
func (m *VMMonitor) Stop() {
	m.logger.Info("Stopping VM monitor")
	close(m.shutdown)

	// Stop all monitoring goroutines
	m.monitoredVMsLock.Lock()
	defer m.monitoredVMsLock.Unlock()

	for _, vm := range m.monitoredVMs {
		if vm.isMonitoring && vm.cancelContext != nil {
			vm.cancelContext()
		}
	}
}

// RegisterVM starts monitoring a VM when clients connect.
func (m *VMMonitor) RegisterVM(vmName string) {
	m.monitoredVMsLock.Lock()
	defer m.monitoredVMsLock.Unlock()

	// Check if VM is already monitored
	vm, exists := m.monitoredVMs[vmName]
	if exists {
		vm.clientCount++
		m.logger.Debug("Incremented client count for monitored VM",
			logger.String("vmName", vmName),
			logger.Int("clientCount", vm.clientCount))
		return
	}

	// Start monitoring new VM
	vm = &monitoredVM{
		name:         vmName,
		lastChecked:  time.Now(),
		stateChanged: time.Now(),
		clientCount:  1,
	}
	m.monitoredVMs[vmName] = vm

	// Get initial VM state
	ctx := context.Background()
	vmInfo, err := m.vmManager.Get(ctx, vmName)
	if err != nil {
		m.logger.Error("Failed to get VM info for monitoring",
			logger.String("vmName", vmName),
			logger.Error(err))
		return
	}

	vm.lastStatus = vmInfo.Status
	if vmInfo.Status == vmmodels.VMStatusRunning {
		vm.startTime = time.Now()
	}

	// Start monitoring goroutine
	m.startMonitoring(vmName)

	m.logger.Info("Started monitoring VM",
		logger.String("vmName", vmName),
		logger.String("status", string(vm.lastStatus)))
}

// UnregisterVM stops monitoring a VM when all clients disconnect.
func (m *VMMonitor) UnregisterVM(vmName string) {
	m.monitoredVMsLock.Lock()
	defer m.monitoredVMsLock.Unlock()

	// Check if VM is monitored
	vm, exists := m.monitoredVMs[vmName]
	if !exists {
		return
	}

	// Decrement client count
	vm.clientCount--

	// If there are still clients, continue monitoring
	if vm.clientCount > 0 {
		m.logger.Debug("Decremented client count for monitored VM",
			logger.String("vmName", vmName),
			logger.Int("clientCount", vm.clientCount))
		return
	}

	// Stop monitoring if no more clients
	if vm.isMonitoring && vm.cancelContext != nil {
		vm.cancelContext()
		vm.isMonitoring = false
	}

	// Remove VM from monitored list
	delete(m.monitoredVMs, vmName)

	m.logger.Info("Stopped monitoring VM",
		logger.String("vmName", vmName))
}

// startMonitoring starts the monitoring goroutine for a VM.
func (m *VMMonitor) startMonitoring(vmName string) {
	m.monitoredVMsLock.Lock()
	vm := m.monitoredVMs[vmName]
	if vm.isMonitoring {
		m.monitoredVMsLock.Unlock()
		return
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	vm.cancelContext = cancel
	vm.isMonitoring = true
	m.monitoredVMsLock.Unlock()

	// Start monitoring goroutine
	go func() {
		ticker := time.NewTicker(VMMonitorInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.collectAndBroadcastMetrics(ctx, vmName)
			}
		}
	}()
}

// collectAndBroadcastMetrics collects VM metrics and broadcasts them.
func (m *VMMonitor) collectAndBroadcastMetrics(ctx context.Context, vmName string) {
	// Get VM status
	vmInfo, err := m.vmManager.Get(ctx, vmName)
	if err != nil {
		m.logger.Error("Failed to get VM status",
			logger.String("vmName", vmName),
			logger.Error(err))
		return
	}

	// Update monitored VM state
	m.monitoredVMsLock.Lock()
	vm, exists := m.monitoredVMs[vmName]
	if !exists {
		m.monitoredVMsLock.Unlock()
		return
	}

	// Check for status change
	statusChanged := vm.lastStatus != vmInfo.Status
	if statusChanged {
		m.logger.Info("VM status changed",
			logger.String("vmName", vmName),
			logger.String("oldStatus", string(vm.lastStatus)),
			logger.String("newStatus", string(vmInfo.Status)))

		vm.lastStatus = vmInfo.Status
		vm.stateChanged = time.Now()

		// Update start time for running VMs
		if vmInfo.Status == vmmodels.VMStatusRunning {
			vm.startTime = time.Now()
		}
	}

	vm.lastChecked = time.Now()
	m.monitoredVMsLock.Unlock()

	// Send status update
	var uptime int64 = 0
	if vmInfo.Status == vmmodels.VMStatusRunning {
		uptime = int64(time.Since(vm.startTime).Seconds())
	}

	m.handler.SendVMStatus(vmName, vmInfo.Status, vm.stateChanged, uptime)

	// Get and send metrics if VM is running
	if vmInfo.Status == vmmodels.VMStatusRunning {
		metrics, err := m.vmManager.GetMetrics(ctx, vmName)
		if err != nil {
			m.logger.Error("Failed to get VM metrics",
				logger.String("vmName", vmName),
				logger.Error(err))
			return
		}

		m.handler.SendVMMetrics(
			vmName,
			metrics.CPU.Utilization,
			metrics.Memory.Used,
			metrics.Memory.Total,
			metrics.Network.RxBytes,
			metrics.Network.TxBytes,
			metrics.Disk.ReadBytes,
			metrics.Disk.WriteBytes,
		)
	}
}

// cleanupRoutine periodically cleans up stale VM monitoring.
func (m *VMMonitor) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.shutdown:
			return
		case <-ticker.C:
			m.cleanupStaleMonitoring()
		}
	}
}

// cleanupStaleMonitoring removes monitoring for VMs that haven't been checked recently.
func (m *VMMonitor) cleanupStaleMonitoring() {
	m.monitoredVMsLock.Lock()
	defer m.monitoredVMsLock.Unlock()

	staleTime := time.Now().Add(-15 * time.Minute)

	for name, vm := range m.monitoredVMs {
		if vm.lastChecked.Before(staleTime) {
			// Stop monitoring goroutine
			if vm.isMonitoring && vm.cancelContext != nil {
				vm.cancelContext()
			}

			// Remove from monitored VMs
			delete(m.monitoredVMs, name)

			m.logger.Info("Cleaned up stale VM monitoring",
				logger.String("vmName", name),
				logger.Time("lastChecked", vm.lastChecked))
		}
	}
}
