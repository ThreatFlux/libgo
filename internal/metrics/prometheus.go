package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/threatflux/libgo/pkg/logger"
)

// PrometheusMetrics implements metrics collection
type PrometheusMetrics struct {
	// HTTP request metrics
	requestDuration *prometheus.HistogramVec
	requests        *prometheus.CounterVec

	// VM operation metrics
	vmOperations *prometheus.CounterVec
	vmCount      prometheus.GaugeFunc

	// Export metrics
	exportCount    prometheus.GaugeFunc
	exportDuration *prometheus.HistogramVec

	// Libvirt metrics
	libvirtErrors  *prometheus.CounterVec
	libvirtLatency *prometheus.HistogramVec

	// Dependencies
	vmManager     interface{}
	exportManager interface{}
	logger        logger.Logger
}

// NewPrometheusMetrics creates a new PrometheusMetrics
func NewPrometheusMetrics(vmManager interface{}, exportManager interface{}, logger logger.Logger) *PrometheusMetrics {
	m := &PrometheusMetrics{
		vmManager:     vmManager,
		exportManager: exportManager,
		logger:        logger,
	}

	// Initialize request metrics
	m.requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	m.requests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"method", "path", "status"},
	)

	// Initialize VM operation metrics
	m.vmOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vm_operations_total",
			Help: "Total number of VM operations",
		},
		[]string{"operation", "vm_name", "status"},
	)

	m.vmCount = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "vm_count",
			Help: "Current number of VMs",
		},
		func() float64 {
			// Type assertion required but might fail in runtime
			// This is a simplified implementation
			m.logger.Error("VM count metrics disabled temporarily")
			return 0
		},
	)

	// Initialize export metrics
	m.exportCount = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "export_jobs_count",
			Help: "Current number of export jobs",
		},
		func() float64 {
			// Type assertion required but might fail in runtime
			// This is a simplified implementation
			m.logger.Error("Export job count metrics disabled temporarily")
			return 0
		},
	)

	m.exportDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "export_duration_seconds",
			Help:    "Duration of export operations in seconds",
			Buckets: []float64{60, 180, 300, 600, 900, 1800, 3600, 7200},
		},
		[]string{"format", "status"},
	)

	// Initialize libvirt metrics
	m.libvirtErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "libvirt_errors_total",
			Help: "Total number of libvirt errors",
		},
		[]string{"operation", "error_type"},
	)

	m.libvirtLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "libvirt_operation_duration_seconds",
			Help:    "Duration of libvirt operations in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10},
		},
		[]string{"operation"},
	)

	return m
}

// RecordRequest records an API request
func (m *PrometheusMetrics) RecordRequest(method, path string, status int, duration time.Duration) {
	statusStr := prometheus.Labels{"method": method, "path": path, "status": string(status)}
	m.requests.With(statusStr).Inc()
	m.requestDuration.With(statusStr).Observe(duration.Seconds())
}

// RecordVMOperation records a VM operation
func (m *PrometheusMetrics) RecordVMOperation(operation string, vmName string, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}

	m.vmOperations.With(prometheus.Labels{
		"operation": operation,
		"vm_name":   vmName,
		"status":    status,
	}).Inc()
}

// RecordExportDuration records the duration of an export operation
func (m *PrometheusMetrics) RecordExportDuration(format string, status string, duration time.Duration) {
	m.exportDuration.With(prometheus.Labels{
		"format": format,
		"status": status,
	}).Observe(duration.Seconds())
}

// RecordLibvirtError records a libvirt error
func (m *PrometheusMetrics) RecordLibvirtError(operation string, errorType string) {
	m.libvirtErrors.With(prometheus.Labels{
		"operation":  operation,
		"error_type": errorType,
	}).Inc()
}

// RecordLibvirtOperation records a libvirt operation latency
func (m *PrometheusMetrics) RecordLibvirtOperation(operation string, duration time.Duration) {
	m.libvirtLatency.With(prometheus.Labels{
		"operation": operation,
	}).Observe(duration.Seconds())
}
