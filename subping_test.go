package subping_test

import (
	"net"
	"testing"
	"time"

	"github.com/fadhilyori/subping"
)

func TestSubping_Run(t *testing.T) {
	targets := []net.IP{
		net.ParseIP("127.0.0.1"),
		net.ParseIP("127.0.0.2"),
		net.ParseIP("128.0.0.0"),
	}
	failedTargetsCount := 1

	opts := &subping.Options{
		Targets: targets,
		Count:   3,
		Timeout: 300 * time.Millisecond,
		NumJobs: 8,
	}

	sp, err := subping.NewSubping(opts)
	if err != nil {
		t.Fatalf("Failed to create Subping instance: %v", err)
	}

	sp.Run()

	results := sp.GetResults()
	onlineResults := sp.GetOnlineHosts()
	targetsCount := len(targets)
	onlineTargetsCount := targetsCount - failedTargetsCount

	if len(results) != targetsCount {
		t.Errorf("Expected %d results, but got %d", len(targets), len(results))
	}

	if len(onlineResults) != onlineTargetsCount {
		t.Errorf("Expected %d results, but got %d", len(targets), len(results))
	}
}

func TestRunPing(t *testing.T) {
	ipAddress := net.ParseIP("127.0.0.1")
	count := 3
	timeout := 300 * time.Millisecond

	stats := subping.RunPing(ipAddress, count, timeout)
	if stats == nil {
		t.Errorf("Failed to get ping statistics for IP address %s", ipAddress)
	}
}
