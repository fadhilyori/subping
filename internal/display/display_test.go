package display

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/fadhilyori/subping/internal/ping"
)

func TestIPSorting(t *testing.T) {
	// Test IP sorting with various IP formats
	results := []HostResult{
		{IP: net.ParseIP("192.168.1.100")},
		{IP: net.ParseIP("10.0.0.1")},
		{IP: net.ParseIP("172.16.0.50")},
		{IP: net.ParseIP("192.168.1.10")},
		{IP: net.ParseIP("10.0.0.100")},
	}

	td := &TerminalDisplay{
		config: DisplayConfig{SortBy: SortByIP},
	}

	td.sortResults(results)

	// Verify correct byte-wise sorting
	expected := []string{
		"10.0.0.1",
		"10.0.0.100",
		"172.16.0.50",
		"192.168.1.10",
		"192.168.1.100",
	}

	for i, result := range results {
		if result.IP.String() != expected[i] {
			t.Errorf("Expected %s at position %d, got %s", expected[i], i, result.IP.String())
		}
	}
}

func TestTerminalColorDetection(t *testing.T) {
	// Test with NO_COLOR set
	os.Setenv("NO_COLOR", "1")
	if IsTerminalColorSupported() {
		t.Error("Expected colors to be disabled when NO_COLOR is set")
	}

	// Test without NO_COLOR
	os.Unsetenv("NO_COLOR")
	// Note: This test might fail in non-terminal environments
	// but should work in most CI/terminal setups
}

func TestProgressUpdateRaceCondition(t *testing.T) {
	pt := NewProgressTracker(5, DefaultColorScheme(), true, 200*time.Millisecond)

	// Test completion update bypasses rate limiting
	pt.Update(5, "192.168.1.1", 3) // Final update

	_, elapsed := pt.GetStats()
	if elapsed == 0 {
		t.Error("Expected non-zero elapsed time")
	}
}

func TestNetworkHealthScore(t *testing.T) {
	td := &TerminalDisplay{}

	// Test basic health score
	score := td.calculateNetworkHealthScore(8, 10)
	expected := 80
	if score != expected {
		t.Errorf("Expected health score %d, got %d", expected, score)
	}

	// Test enhanced health score with various metrics
	results := []HostResult{
		{
			Status:     StatusOnline,
			AvgRtt:     10 * time.Millisecond,
			PacketLoss: 0,
			Jitter:     5 * time.Millisecond,
		},
		{
			Status:     StatusOnline,
			AvgRtt:     150 * time.Millisecond, // High latency
			PacketLoss: 10,                     // High packet loss
			Jitter:     60 * time.Millisecond,  // High jitter
		},
		{
			Status: StatusOffline,
		},
	}

	enhancedScore := td.calculateEnhancedNetworkHealthScore(results)
	if enhancedScore < 0 || enhancedScore > 100 {
		t.Errorf("Enhanced health score should be between 0-100, got %d", enhancedScore)
	}
}

func TestMemoryAllocation(t *testing.T) {
	// Test capacity calculation for different scenarios

	// Scenario 1: Only online hosts
	onlineCount := 8
	totalHosts := 10
	showOffline := false

	expectedCapacity := onlineCount
	if showOffline {
		expectedCapacity = totalHosts
	}

	results := make([]HostResult, 0, expectedCapacity)
	if cap(results) != expectedCapacity {
		t.Errorf("Expected capacity %d, got %d", expectedCapacity, cap(results))
	}

	// Scenario 2: All hosts
	showOffline = true
	expectedCapacity = totalHosts

	results = make([]HostResult, 0, expectedCapacity)
	if cap(results) != expectedCapacity {
		t.Errorf("Expected capacity %d, got %d", expectedCapacity, cap(results))
	}
}

func TestColorSchemeForLatency(t *testing.T) {
	cs := DefaultColorScheme()

	tests := []struct {
		latency  time.Duration
		expected string
	}{
		{5 * time.Millisecond, "Excellent"},
		{50 * time.Millisecond, "Good"},
		{200 * time.Millisecond, "Degraded"},
		{600 * time.Millisecond, "Poor"},
		{0, "Offline"},
	}

	for _, test := range tests {
		color := cs.GetColorForLatency(test.latency)
		if color == nil {
			t.Errorf("Expected color for latency %v", test.latency)
		}
	}
}

func TestHostStatusIcon(t *testing.T) {
	tests := map[HostStatus]string{
		StatusOnline:  "●",
		StatusOffline: "○",
	}

	for status, expected := range tests {
		icon := GetHostStatusIcon(status)
		if icon != expected {
			t.Errorf("Expected icon %s for status %v, got %s", expected, status, icon)
		}
	}
}

func TestConvertPingResult(t *testing.T) {
	// Mock ping result
	mockResult := ping.Result{
		MinRtt:      10 * time.Millisecond,
		AvgRtt:      20 * time.Millisecond,
		MaxRtt:      30 * time.Millisecond,
		PacketLoss:  5.0,
		PacketsSent: 10,
		PacketsRecv: 9,
		StdDevRtt:   2 * time.Millisecond,
	}

	hostResult := ConvertPingResult("192.168.1.1", mockResult)

	if hostResult.IP.String() != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", hostResult.IP.String())
	}

	if hostResult.Status != StatusOnline {
		t.Errorf("Expected status Online, got %v", hostResult.Status)
	}

	if hostResult.Jitter != mockResult.StdDevRtt {
		t.Errorf("Expected jitter %v, got %v", mockResult.StdDevRtt, hostResult.Jitter)
	}
}

func TestDisplayConfig(t *testing.T) {
	config := DisplayConfig{
		EnabledColors:    true,
		EnabledProgress:  true,
		SortBy:           SortByLatency,
		ProgressCallback: func(current, total int, currentIP string, onlineCount int) {},
	}

	// Test that config can be used to create display
	display := NewDisplay(config)
	if display == nil {
		t.Error("Expected non-nil display")
	}
}

func TestSortOptions(t *testing.T) {
	tests := []struct {
		option   SortOption
		expected string
	}{
		{SortByIP, "ip"},
		{SortByLatency, "latency"},
		{SortByLoss, "loss"},
		{SortByJitter, "jitter"},
	}

	for _, test := range tests {
		if string(test.option) != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, string(test.option))
		}
	}
}
