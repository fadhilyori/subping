package display

import (
	"net"
	"testing"
	"time"

	"github.com/fadhilyori/subping/internal/ping"
)

func TestDisplayPipelineIntegration(t *testing.T) {
	// Test complete display pipeline with mock data
	config := DisplayConfig{
		EnabledColors:   false, // Disable colors for consistent testing
		EnabledProgress: false, // Disable progress for cleaner test output
		SortBy:          SortByIP,
	}

	display := NewDisplay(config)

	// Test header display
	display.ShowHeader(
		"192.168.1.0/24",
		"192.168.1.1 - 192.168.1.254",
		254,
		32,
		3,
		"100ms",
		"500ms",
		"test-version",
	)

	// Create mock results
	results := []HostResult{
		{
			IP:          net.ParseIP("192.168.1.1"),
			Status:      StatusOnline,
			MinRtt:      5 * time.Millisecond,
			AvgRtt:      10 * time.Millisecond,
			MaxRtt:      15 * time.Millisecond,
			PacketLoss:  0,
			Jitter:      2 * time.Millisecond,
			PacketsSent: 3,
			PacketsRecv: 3,
		},
		{
			IP:          net.ParseIP("192.168.1.2"),
			Status:      StatusOffline,
			MinRtt:      0,
			AvgRtt:      0,
			MaxRtt:      0,
			PacketLoss:  100,
			Jitter:      0,
			PacketsSent: 3,
			PacketsRecv: 0,
		},
		{
			IP:          net.ParseIP("192.168.1.10"),
			Status:      StatusOnline,
			MinRtt:      50 * time.Millisecond,
			AvgRtt:      75 * time.Millisecond,
			MaxRtt:      100 * time.Millisecond,
			PacketLoss:  10,
			Jitter:      25 * time.Millisecond,
			PacketsSent: 3,
			PacketsRecv: 2,
		},
	}

	// Test results display
	display.ShowResults(results)

	// Test summary display
	display.ShowSummary(3, 2, 1, 5*time.Second, 0.6)
}

func TestProgressCallbackIntegration(t *testing.T) {
	var progressCalls []struct {
		current     int
		total       int
		currentIP   string
		onlineCount int
	}

	config := DisplayConfig{
		EnabledColors:   false,
		EnabledProgress: true,
		SortBy:          SortByIP,
		ProgressCallback: func(current, total int, currentIP string, onlineCount int) {
			progressCalls = append(progressCalls, struct {
				current     int
				total       int
				currentIP   string
				onlineCount int
			}{current, total, currentIP, onlineCount})
		},
	}

	display := NewDisplay(config)

	// Simulate progress updates
	display.UpdateProgress(1, 10, "192.168.1.1", 1)
	display.UpdateProgress(5, 10, "192.168.1.5", 3)
	display.UpdateProgress(10, 10, "192.168.1.10", 7)

	if len(progressCalls) != 3 {
		t.Errorf("Expected 3 progress calls, got %d", len(progressCalls))
	}

	// Verify last call
	lastCall := progressCalls[len(progressCalls)-1]
	if lastCall.current != 10 || lastCall.total != 10 {
		t.Error("Progress callback not called with correct final values")
	}
}

func TestColorOutputIntegration(t *testing.T) {
	// Test color configuration
	config := DisplayConfig{
		EnabledColors:   true,
		EnabledProgress: false,
		SortBy:          SortByIP,
	}

	display := NewDisplay(config)

	// Create test results
	results := []HostResult{
		{
			IP:     net.ParseIP("192.168.1.1"),
			Status: StatusOnline,
			AvgRtt: 10 * time.Millisecond,
		},
	}

	// Test that display doesn't panic with colors enabled
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Display panicked with colors: %v", r)
		}
	}()

	display.ShowResults(results)
}

func TestSortingIntegration(t *testing.T) {
	// Test all sorting options
	testResults := []HostResult{
		{
			IP:         net.ParseIP("192.168.1.100"),
			Status:     StatusOnline,
			AvgRtt:     100 * time.Millisecond,
			PacketLoss: 5,
			Jitter:     10 * time.Millisecond,
		},
		{
			IP:         net.ParseIP("192.168.1.1"),
			Status:     StatusOnline,
			AvgRtt:     10 * time.Millisecond,
			PacketLoss: 0,
			Jitter:     2 * time.Millisecond,
		},
		{
			IP:         net.ParseIP("192.168.1.50"),
			Status:     StatusOnline,
			AvgRtt:     50 * time.Millisecond,
			PacketLoss: 10,
			Jitter:     20 * time.Millisecond,
		},
	}

	sortOptions := []SortOption{SortByIP, SortByLatency, SortByLoss, SortByJitter}

	for _, sortBy := range sortOptions {
		config := DisplayConfig{
			EnabledColors:   false,
			EnabledProgress: false,
			SortBy:          sortBy,
		}

		display := NewDisplay(config)
		terminalDisplay := display.(*TerminalDisplay)

		// Copy results to avoid modifying original
		testCopy := make([]HostResult, len(testResults))
		copy(testCopy, testResults)

		terminalDisplay.ShowResults(testCopy)

		// Verify sorting worked (basic check)
		if len(testCopy) != len(testResults) {
			t.Errorf("Sorting with %s modified result count", sortBy)
		}
	}
}

func TestErrorHandlingIntegration(t *testing.T) {
	// Test display behavior with edge cases
	config := DisplayConfig{
		EnabledColors:   false,
		EnabledProgress: false,
		SortBy:          SortByIP,
	}

	display := NewDisplay(config)

	// Test with empty results
	display.ShowResults([]HostResult{})

	// Test with nil results (should not panic)
	display.ShowResults(nil)

	// Test summary with zero values
	display.ShowSummary(0, 0, 0, 0, 0)

	// Test with very large execution time
	display.ShowSummary(100, 80, 20, 3600*time.Second, 0.0278)
}

func TestPerformanceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Test display performance with large result sets
	config := DisplayConfig{
		EnabledColors:   false,
		EnabledProgress: false,
		SortBy:          SortByIP,
	}

	display := NewDisplay(config)

	// Generate large result set
	results := make([]HostResult, 1000)
	for i := 0; i < 1000; i++ {
		ip := net.IPv4(192, 168, 1, byte(i%254+1))
		results[i] = HostResult{
			IP:         ip,
			Status:     StatusOnline,
			AvgRtt:     time.Duration(i%100) * time.Millisecond,
			PacketLoss: float64(i % 20),
			Jitter:     time.Duration(i%10) * time.Millisecond,
		}
	}

	start := time.Now()
	display.ShowResults(results)
	duration := time.Since(start)

	// Should complete within reasonable time (adjust threshold as needed)
	if duration > 5*time.Second {
		t.Errorf("Display took too long: %v", duration)
	}
}

func TestConvertPingResultIntegration(t *testing.T) {
	// Test conversion with various ping result scenarios
	testCases := []struct {
		name       string
		pingResult ping.Result
		ip         string
		expected   HostResult
	}{
		{
			name: "online host",
			pingResult: ping.Result{
				MinRtt:      5 * time.Millisecond,
				AvgRtt:      10 * time.Millisecond,
				MaxRtt:      15 * time.Millisecond,
				PacketLoss:  0,
				PacketsSent: 3,
				PacketsRecv: 3,
				StdDevRtt:   2 * time.Millisecond,
			},
			ip: "192.168.1.1",
			expected: HostResult{
				Status:      StatusOnline,
				MinRtt:      5 * time.Millisecond,
				AvgRtt:      10 * time.Millisecond,
				MaxRtt:      15 * time.Millisecond,
				PacketLoss:  0,
				PacketsSent: 3,
				PacketsRecv: 3,
				Jitter:      2 * time.Millisecond,
			},
		},
		{
			name: "offline host",
			pingResult: ping.Result{
				MinRtt:      0,
				AvgRtt:      0,
				MaxRtt:      0,
				PacketLoss:  100,
				PacketsSent: 3,
				PacketsRecv: 0,
				StdDevRtt:   0,
			},
			ip: "192.168.1.2",
			expected: HostResult{
				Status:      StatusOffline,
				MinRtt:      0,
				AvgRtt:      0,
				MaxRtt:      0,
				PacketLoss:  100,
				PacketsSent: 3,
				PacketsRecv: 0,
				Jitter:      0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertPingResult(tc.ip, tc.pingResult)

			if result.Status != tc.expected.Status {
				t.Errorf("Expected status %v, got %v", tc.expected.Status, result.Status)
			}

			if result.IP.String() != tc.ip {
				t.Errorf("Expected IP %s, got %s", tc.ip, result.IP.String())
			}
		})
	}
}
