package metrics

import (
	"context"
	"time"

	"github.com/threatflux/libgo/pkg/logger"
)

// Collector provides an interface for metrics collection
type Collector interface {
	// RecordRequest records a HTTP request
	RecordRequest(method, path string, status int, duration time.Duration)

	// RecordVMOperation records a VM operation
	RecordVMOperation(operation string, vmName string, success bool)

	// RecordExportDuration records the duration of an export operation
	RecordExportDuration(format string, status string, duration time.Duration)

	// RecordLibvirtError records a libvirt error
	RecordLibvirtError(operation string, errorType string)

	// RecordLibvirtOperation records a libvirt operation latency
	RecordLibvirtOperation(operation string, duration time.Duration)
}

// NewCollector creates a new metrics collector
func NewCollector(impl string, ctx context.Context, deps map[string]interface{}, logger logger.Logger) (Collector, error) {
	switch impl {
	case "prometheus":
		// Get dependencies from the map
		vmManager, ok := deps["vm_manager"]
		if !ok {
			logger.Error("VM manager dependency not provided")
			return nil, nil
		}

		exportManager, ok := deps["export_manager"]
		if !ok {
			logger.Error("Export manager dependency not provided")
			return nil, nil
		}

		return NewPrometheusMetrics(vmManager, exportManager, logger), nil

	case "noop":
		return &NoopCollector{}, nil

	default:
		return &NoopCollector{}, nil
	}
}

// NoopCollector is a no-operation metrics collector for testing or when metrics are disabled
type NoopCollector struct{}

// RecordRequest is a no-op implementation
func (n *NoopCollector) RecordRequest(method, path string, status int, duration time.Duration) {}

// RecordVMOperation is a no-op implementation
func (n *NoopCollector) RecordVMOperation(operation string, vmName string, success bool) {}

// RecordExportDuration is a no-op implementation
func (n *NoopCollector) RecordExportDuration(format string, status string, duration time.Duration) {}

// RecordLibvirtError is a no-op implementation
func (n *NoopCollector) RecordLibvirtError(operation string, errorType string) {}

// RecordLibvirtOperation is a no-op implementation
func (n *NoopCollector) RecordLibvirtOperation(operation string, duration time.Duration) {}
