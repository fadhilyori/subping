package ping

import (
	"fmt"
	"runtime"
	"time"

	ping "github.com/prometheus-community/pro-bing"
)

// noopLogger is a logger that discards all output
type noopLogger struct{}

func (l *noopLogger) Debugf(format string, v ...interface{}) {}
func (l *noopLogger) Infof(format string, v ...interface{})  {}
func (l *noopLogger) Warnf(format string, v ...interface{})  {}
func (l *noopLogger) Errorf(format string, v ...interface{}) {}
func (l *noopLogger) Fatalf(format string, v ...interface{}) {}
func (l *noopLogger) Debug(v ...interface{})                 {}
func (l *noopLogger) Info(v ...interface{})                  {}
func (l *noopLogger) Warn(v ...interface{})                  {}
func (l *noopLogger) Error(v ...interface{})                 {}
func (l *noopLogger) Fatal(v ...interface{})                 {}

// realPinger is the production implementation using pro-bing library
// It performs actual ICMP ping operations
type realPinger struct{}

// NewRealPinger creates a new real pinger instance
func NewRealPinger() Pinger {
	return &realPinger{}
}

// Ping implements the Pinger interface using the pro-bing library
func (p *realPinger) Ping(ipAddress string, count int, interval time.Duration, timeout time.Duration) (Result, error) {
	// Create a new pinger for the target address
	pinger, err := ping.NewPinger(ipAddress)
	if err != nil {
		return Result{}, err
	}

	// Configure pinger parameters
	pinger.Count = count
	pinger.Interval = interval

	if timeout > 0 {
		pinger.Timeout = timeout
	}

	// Windows requires privileged mode for ICMP operations
	if runtime.GOOS == "windows" {
		pinger.SetPrivileged(true)
	}

	// Use a custom logger that discards output to prevent log noise in progress bar
	// This creates a no-op logger that implements the required interface
	pinger.SetLogger(&noopLogger{})

	// Execute the ping operation
	err = pinger.Run()
	if err != nil {
		return Result{}, err
	}

	// Get the statistics and convert to our Result type
	stats := pinger.Statistics()
	if stats == nil {
		return Result{}, fmt.Errorf("failed to get ping statistics for %s", ipAddress)
	}

	return Result{
		AvgRtt:                stats.AvgRtt,
		PacketLoss:            stats.PacketLoss,
		PacketsSent:           stats.PacketsSent,
		PacketsRecv:           stats.PacketsRecv,
		PacketsRecvDuplicates: stats.PacketsRecvDuplicates,
		MinRtt:                stats.MinRtt,
		MaxRtt:                stats.MaxRtt,
		StdDevRtt:             stats.StdDevRtt,
	}, nil
}
