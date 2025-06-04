package health

import (
	"runtime"
	"sync"
	"time"
)

// Status represents health status.
type Status string

// Health status constants.
const (
	StatusUp   Status = "UP"
	StatusDown Status = "DOWN"
)

// Check represents a health check.
type Check struct {
	Details map[string]string `json:"details,omitempty"`
	Name    string            `json:"name"`
	Status  Status            `json:"status"`
}

// Result represents health check result.
type Result struct {
	Status    Status  `json:"status"`
	Version   string  `json:"version"`
	BuildTime string  `json:"buildTime,omitempty"`
	GoVersion string  `json:"goVersion"`
	GOOS      string  `json:"os"`
	GOARCH    string  `json:"arch"`
	Uptime    string  `json:"uptime"`
	Checks    []Check `json:"checks"`
}

// CheckFunction represents a health check function.
type CheckFunction func() Check

// Checker performs health checks.
type Checker struct {
	startTime time.Time
	version   string
	buildTime string
	checks    []CheckFunction
	mu        sync.RWMutex
}

// NewChecker creates a new health Checker.
func NewChecker(version, buildTime string) *Checker {
	return &Checker{
		checks:    make([]CheckFunction, 0),
		version:   version,
		buildTime: buildTime,
		startTime: time.Now(),
	}
}

// AddCheck adds a health check.
func (c *Checker) AddCheck(check CheckFunction) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks = append(c.checks, check)
}

// RunChecks runs all health checks.
func (c *Checker) RunChecks() Result {
	c.mu.RLock()
	checksToRun := make([]CheckFunction, len(c.checks))
	copy(checksToRun, c.checks)
	c.mu.RUnlock()

	// Run all checks
	checkResults := make([]Check, 0, len(checksToRun))
	overallStatus := StatusUp

	for _, checkFn := range checksToRun {
		check := checkFn()
		checkResults = append(checkResults, check)

		// If any check is down, overall status is down
		if check.Status == StatusDown {
			overallStatus = StatusDown
		}
	}

	// Calculate uptime
	uptime := time.Since(c.startTime).String()

	// Return health check result
	return Result{
		Status:    overallStatus,
		Checks:    checkResults,
		Version:   c.version,
		BuildTime: c.buildTime,
		GoVersion: runtime.Version(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		Uptime:    uptime,
	}
}
