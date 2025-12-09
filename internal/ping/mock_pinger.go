package ping

import (
	"errors"
	"net"
	"sync"
	"time"
)

// MockPingerConfig defines the behavior of the mock pinger
type MockPingerConfig struct {
	// DefaultLatency is the default round-trip time for successful pings
	DefaultLatency time.Duration

	// DefaultPacketLoss is the default packet loss percentage (0.0 to 1.0)
	DefaultPacketLoss float64

	// HostConfigs allows configuring specific behavior for different hosts
	HostConfigs map[string]MockHostConfig

	// SimulateTiming adds realistic delay to simulate network latency
	SimulateTiming bool
}

// MockHostConfig defines specific behavior for a particular host
type MockHostConfig struct {
	// Latency is the round-trip time for this host
	Latency time.Duration

	// PacketLoss is the packet loss percentage for this host (0.0 to 1.0)
	PacketLoss float64

	// ShouldError forces the ping to fail with an error
	ShouldError bool

	// ErrorMsg is the error message to return when ShouldError is true
	ErrorMsg string
}

// mockPinger is a mock implementation for testing purposes
// It provides deterministic ping results without requiring network access
type mockPinger struct {
	config MockPingerConfig
	mu     sync.RWMutex
}

// NewMockPinger creates a new mock pinger with default configuration
func NewMockPinger() Pinger {
	return NewMockPingerWithConfig(MockPingerConfig{
		DefaultLatency:     10 * time.Millisecond,
		DefaultPacketLoss: 0.0,
		HostConfigs:       make(map[string]MockHostConfig),
		SimulateTiming:     true,
	})
}

// NewMockPingerWithConfig creates a new mock pinger with custom configuration
func NewMockPingerWithConfig(config MockPingerConfig) Pinger {
	return &mockPinger{
		config: config,
	}
}

// Ping implements the Pinger interface with mock behavior
func (p *mockPinger) Ping(ipAddress string, count int, interval time.Duration, timeout time.Duration) (Result, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Simulate network delay if configured
	if p.config.SimulateTiming {
		time.Sleep(1 * time.Millisecond) // Minimal delay to simulate network operation
	}

	// Validate IP address first
	ip := net.ParseIP(ipAddress)
	if ip == nil && ipAddress != "localhost" {
		// Invalid IP address should return an error
		return Result{}, errors.New("invalid IP address")
	}

	// Check for host-specific configuration first
	if hostConfig, exists := p.config.HostConfigs[ipAddress]; exists {
		if hostConfig.ShouldError {
			return Result{}, errors.New(hostConfig.ErrorMsg)
		}

		return p.calculateResult(count, hostConfig.Latency, hostConfig.PacketLoss), nil
	}

	// Default behavior based on IP address patterns
	if p.isLocalhost(ipAddress) {
		// Localhost should respond quickly and reliably
		return p.calculateResult(count, 1*time.Millisecond, 0.0), nil
	}

	if p.isPrivateIP(ipAddress) {
		// Private IPs should respond with moderate latency and low packet loss
		return p.calculateResult(count, p.config.DefaultLatency, p.config.DefaultPacketLoss), nil
	}

	// Public/external IPs get higher latency and some packet loss to simulate real conditions
	return p.calculateResult(count, 50*time.Millisecond, 0.1), nil
}

// SetHostConfig allows updating configuration for specific hosts (useful for testing)
func (p *mockPinger) SetHostConfig(ipAddress string, config MockHostConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.config.HostConfigs == nil {
		p.config.HostConfigs = make(map[string]MockHostConfig)
	}
	p.config.HostConfigs[ipAddress] = config
}

// calculateResult generates ping statistics based on input parameters
func (p *mockPinger) calculateResult(count int, latency time.Duration, packetLoss float64) Result {
	packetsRecv := int(float64(count) * (1.0 - packetLoss))

	// Ensure at least 0 packets received
	if packetsRecv < 0 {
		packetsRecv = 0
	}

	// Convert packet loss from ratio to percentage
	packetLossPercentage := packetLoss * 100.0

	return Result{
		AvgRtt:                latency,
		PacketLoss:            packetLossPercentage,
		PacketsSent:           count,
		PacketsRecv:           packetsRecv,
		PacketsRecvDuplicates: 0, // Mock doesn't simulate duplicates by default
	}
}

// isLocalhost checks if the IP address is localhost
func (p *mockPinger) isLocalhost(ipAddress string) bool {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return ipAddress == "localhost"
	}

	// Check for localhost equivalents
	return ipAddress == "localhost" ||
		   ipAddress == "127.0.0.1" ||
		   ipAddress == "::1" ||
		   ip.IsLoopback()
}

// isPrivateIP checks if the IP address is in a private range
func (p *mockPinger) isPrivateIP(ipAddress string) bool {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false
	}

	// Check for private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",        // RFC 1918
		"172.16.0.0/12",     // RFC 1918
		"192.168.0.0/16",    // RFC 1918
		"fc00::/7",          // IPv6 Unique Local Addresses
		"fe80::/10",         // IPv6 Link-Local Addresses
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network != nil && network.Contains(ip) {
			return true
		}
	}

	return false
}

// ClearHostConfigs removes all host-specific configurations
func (p *mockPinger) ClearHostConfigs() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config.HostConfigs = make(map[string]MockHostConfig)
}

// GetHostConfig returns the configuration for a specific host
func (p *mockPinger) GetHostConfig(ipAddress string) (MockHostConfig, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	config, exists := p.config.HostConfigs[ipAddress]
	return config, exists
}