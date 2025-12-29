// Package display provides terminal output formatting and progress tracking
// for network scanning results. It supports colored output, sortable tables,
// and real-time progress updates with various display configurations.
package display

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

type ProgressTracker struct {
	bar         *progressbar.ProgressBar
	startTime   time.Time
	lastUpdate  time.Time
	totalHosts  int
	onlineCount int
	currentIP   string
	colorScheme *ColorScheme
	enabled     bool
	completed   bool
	throttle    time.Duration
	mu          sync.Mutex
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(totalHosts int, colorScheme *ColorScheme, enabled bool, throttle time.Duration) *ProgressTracker {
	pt := &ProgressTracker{
		startTime:   time.Now(),
		totalHosts:  totalHosts,
		colorScheme: colorScheme,
		enabled:     enabled,
		lastUpdate:  time.Now(),
		throttle:    throttle,
	}

	if enabled {
		pt.bar = progressbar.NewOptions64(
			int64(totalHosts),
			progressbar.OptionSetDescription("Scanning"),
			progressbar.OptionSetWriter(os.Stderr), // Use stderr for progress bar
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("hosts"),
			progressbar.OptionOnCompletion(func() {
				// Don't print anything here, let Finish() handle it
			}),
			progressbar.OptionThrottle(throttle),
			progressbar.OptionFullWidth(),
			progressbar.OptionSetRenderBlankState(false),
		)
	}

	return pt
}

// Update updates the progress bar with current status
func (pt *ProgressTracker) Update(current int, currentIP string, onlineCount int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.currentIP = currentIP
	pt.onlineCount = onlineCount

	if !pt.enabled || pt.bar == nil {
		return
	}

	// Always process completion updates to avoid race condition
	isCompleted := current == pt.totalHosts
	if isCompleted {
		pt.completed = true
	}

	if !isCompleted && time.Since(pt.lastUpdate) < pt.throttle {
		return
	}

	pt.lastUpdate = time.Now()

	_ = pt.bar.Set64(int64(current))

	if isCompleted {
		elapsed := time.Since(pt.startTime)
		rate := float64(current) / elapsed.Seconds()
		pt.printStats(rate, 0)
	}
}

// printStats prints additional progress statistics
func (pt *ProgressTracker) printStats(rate float64, eta time.Duration) {
	if !pt.enabled {
		return
	}

	var rateStr string
	if rate < 1 {
		rateStr = fmt.Sprintf("%.1f hosts/min", rate*60)
	} else {
		rateStr = fmt.Sprintf("%.1f hosts/sec", rate)
	}

	etaStr := "0s"
	if eta > 0 && eta < time.Hour {
		etaStr = fmt.Sprintf("%.0fs", eta.Seconds())
	} else if eta >= time.Hour {
		etaStr = fmt.Sprintf("%.0fm", eta.Minutes())
	}

	stats := fmt.Sprintf(
		"Rate: %s | Online: %d | ETA: %s",
		rateStr,
		pt.onlineCount,
		etaStr,
	)

	fmt.Fprint(os.Stderr, "\r\033[K") // Clear line with ANSI escape sequence
	if pt.currentIP != "" {
		stats += " | Current: " + pt.currentIP
	}
	pt.colorScheme.Progress.Fprintf(os.Stderr, "%s\n", stats)
}

// Finish marks the progress as complete
func (pt *ProgressTracker) Finish() {
	if pt.enabled && pt.bar != nil {
		_ = pt.bar.Finish()
		fmt.Fprintln(os.Stderr)
	}
}

// GetStats returns current progress statistics
func (pt *ProgressTracker) GetStats() (rate float64, elapsed time.Duration) {
	elapsed = time.Since(pt.startTime)
	// Add minimum elapsed time threshold to avoid division by very small numbers
	if elapsed.Seconds() > 0.01 { // 10ms minimum
		rate = float64(pt.totalHosts) / elapsed.Seconds()
	}
	return
}
