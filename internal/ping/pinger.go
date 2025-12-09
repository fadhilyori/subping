package ping

import (
	"os"
	"time"
)

// Result represents the statistics from a ping operation
type Result struct {
	AvgRtt                time.Duration // Average round-trip time
	PacketLoss            float64       // Packet loss percentage
	PacketsSent           int           // Number of packets sent
	PacketsRecv           int           // Number of packets received
	PacketsRecvDuplicates int           // Number of duplicate packets received
}

// Statistics represents the full ping statistics, compatible with pro-bing.Statistics
type Statistics struct {
	// PacketsRecv is the number of packets received.
	PacketsRecv int

	// PacketsSent is the number of packets sent.
	PacketsSent int

	// PacketsRecvDuplicates is the number of duplicate responses there were to a sent packet.
	PacketsRecvDuplicates int

	// PacketLoss is the percentage of packets lost.
	PacketLoss float64

	// AvgRtt is the average round-trip time sent via this pinger.
	AvgRtt time.Duration

	// MinRtt is the minimum round-trip time sent via this pinger.
	MinRtt time.Duration

	// MaxRtt is the maximum round-trip time sent via this pinger.
	MaxRtt time.Duration

	// StdDevRtt is the standard deviation of the round-trip times sent via this pinger.
	StdDevRtt time.Duration
}

// Pinger defines the interface for ping operations
// This allows us to inject different implementations (real or mock) for testing
type Pinger interface {
	// Ping performs a ping operation on the given IP address and returns statistics
	Ping(ipAddress string, count int, interval time.Duration, timeout time.Duration) (Result, error)
}

// NewPinger creates a new pinger instance based on the environment
// In CI/GitHub Actions, it returns a mock pinger to avoid permission issues
// Otherwise, it returns a real pinger for actual ICMP operations
func NewPinger() Pinger {
	// Check if we're in a CI environment
	if isCIEnvironment() {
		return NewMockPinger()
	}
	return NewRealPinger()
}

// NewPingerWithOptions creates a pinger with explicit type selection
// This allows forcing a specific pinger type for testing or special scenarios
func NewPingerWithOptions(pingerType string) Pinger {
	switch pingerType {
	case "mock":
		return NewMockPinger()
	case "real":
		return NewRealPinger()
	default:
		return NewPinger() // auto-detect
	}
}

// isCIEnvironment detects if we're running in a CI environment
// It checks common environment variables set by CI systems
func isCIEnvironment() bool {
	// Check common CI environment variables
	ciVars := []string{
		"CI",                    // Generic CI (set by GitHub Actions, Travis, etc.)
		"GITHUB_ACTIONS",        // GitHub Actions specific
		"CONTINUOUS_INTEGRATION", // Generic CI
		"TRAVIS",               // Travis CI
		"CIRCLECI",             // CircleCI
		"JENKINS_URL",          // Jenkins
		"GITLAB_CI",            // GitLab CI
		"APPVEYOR",             // AppVeyor
		"CI_NAME",              // Various CI systems
		"BUILDKITE",            // Buildkite
		"SEMAPHORE",            // Semaphore CI
	}

	for _, v := range ciVars {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return false
}

// RunPing performs a ping operation to the specified IP address.
// This is a utility function that provides the same interface as the original
// standalone RunPing function but uses our internal pinger abstraction.
// It sends the specified number of ping requests with the given interval and timeout.
func RunPing(ipAddress string, count int, interval time.Duration, timeout time.Duration) Statistics {
	// Use the real pinger for this utility function
	pinger := NewRealPinger()

	result, err := pinger.Ping(ipAddress, count, interval, timeout)
	if err != nil {
		// Return empty statistics on error to maintain compatibility
		return Statistics{}
	}

	// Convert our internal Result to Statistics for compatibility
	return Statistics{
		PacketsSent:           result.PacketsSent,
		PacketsRecv:           result.PacketsRecv,
		PacketsRecvDuplicates: result.PacketsRecvDuplicates,
		PacketLoss:            result.PacketLoss,
		AvgRtt:               result.AvgRtt,
		// Note: We don't track individual RTTs in our Result
		// So min/max/stddev will be zero-initialized, which is acceptable for compatibility
		MinRtt:               0,
		MaxRtt:               0,
		StdDevRtt:            0,
	}
}

