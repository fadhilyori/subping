// Package display provides terminal output formatting and progress tracking
// for network scanning results. It supports colored output, sortable tables,
// and real-time progress updates with various display configurations.
package display

import (
	"net"
	"time"

	"github.com/fadhilyori/subping/internal/ping"
)

// SortOption defines how results can be sorted
type SortOption string

const (
	SortByIP      SortOption = "ip"
	SortByLatency SortOption = "latency"
	SortByLoss    SortOption = "loss"
	SortByJitter  SortOption = "jitter"
)

// DisplayConfig holds configuration for display
type DisplayConfig struct {
	EnabledColors          bool
	EnabledProgress        bool
	SortBy                 SortOption
	ProgressCallback       func(current, total int, currentIP string, onlineCount int)
	UseEnhancedHealthScore bool
	progressThrottle       time.Duration // Internal field
}

// HostResult represents a host's ping result with additional metrics
type HostResult struct {
	IP          net.IP
	Status      HostStatus
	MinRtt      time.Duration
	AvgRtt      time.Duration
	MaxRtt      time.Duration
	PacketLoss  float64
	Jitter      time.Duration
	PacketsSent int
	PacketsRecv int
}

// HostStatus represents the status of a host
type HostStatus int

const (
	StatusOffline HostStatus = iota
	StatusOnline
)

// ProgressReporter handles progress updates during scanning
type ProgressReporter interface {
	UpdateProgress(current, total int, currentIP string, onlineCount int)
}

// ResultDisplayer handles the display of scan results and summaries
type ResultDisplayer interface {
	ShowResults(results []HostResult)
	ShowSummary(totalHosts int, onlineHosts int, offlineHosts int, executionTime time.Duration)
}

// Display interface combines all display functionality for different output formats
type Display interface {
	ProgressReporter
	ResultDisplayer
	ShowHeader(network string, ipRange string, totalHosts int, workers int, count int, interval string, timeout string, version string)
}

// NewDisplay creates a new display instance
func NewDisplay(config DisplayConfig) Display {
	return NewTerminalDisplay(config)
}

// ConvertPingResult converts internal ping.Result to HostResult
func ConvertPingResult(ip string, result ping.Result) HostResult {
	hostIP := net.ParseIP(ip)
	status := StatusOffline
	if result.PacketsRecv > 0 {
		status = StatusOnline
	}

	return HostResult{
		IP:          hostIP,
		Status:      status,
		MinRtt:      result.MinRtt,
		AvgRtt:      result.AvgRtt,
		MaxRtt:      result.MaxRtt,
		PacketLoss:  result.PacketLoss,
		Jitter:      calculateJitter(result),
		PacketsSent: result.PacketsSent,
		PacketsRecv: result.PacketsRecv,
	}
}

// calculateJitter calculates jitter from ping result
func calculateJitter(result ping.Result) time.Duration {
	// Use the actual standard deviation if available
	if result.StdDevRtt > 0 {
		return result.StdDevRtt
	}

	// Fallback to approximation for older implementations
	if result.PacketsRecv == 0 {
		return 0
	}
	// Approximate jitter as 10% of average RTT for responsive hosts
	return time.Duration(float64(result.AvgRtt.Nanoseconds()) * 0.1)
}
