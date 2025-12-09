package ping

import (
	"runtime"
	"time"

	ping "github.com/prometheus-community/pro-bing"
	"github.com/sirupsen/logrus"
)

// realPinger is the production implementation using pro-bing library
// It performs actual ICMP ping operations
type realPinger struct{}

// NewRealPinger creates a new real pinger instance
func NewRealPinger() Pinger {
	return &realPinger{}
}

// Ping implements the Pinger interface using the pro-bing library
// This performs actual network ping operations and returns real statistics
func (p *realPinger) Ping(ipAddress string, count int, interval time.Duration, timeout time.Duration) (Result, error) {
	// Create a new pinger for the target address
	pinger, err := ping.NewPinger(ipAddress)
	if err != nil {
		logrus.Printf("Failed to create pinger for IP Address: %s\n", ipAddress)
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

	// Execute the ping operation
	err = pinger.Run()
	if err != nil {
		logrus.Printf("Failed to ping the address %s, %v\n", ipAddress, err.Error())
		return Result{}, err
	}

	// Get the statistics and convert to our Result type
	stats := pinger.Statistics()
	return Result{
		AvgRtt:                stats.AvgRtt,
		PacketLoss:            stats.PacketLoss,
		PacketsSent:           stats.PacketsSent,
		PacketsRecv:           stats.PacketsRecv,
		PacketsRecvDuplicates: stats.PacketsRecvDuplicates,
	}, nil
}