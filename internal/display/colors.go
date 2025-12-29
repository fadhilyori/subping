// Package display provides terminal output formatting and progress tracking
// for network scanning results. It supports colored output, sortable tables,
// and real-time progress updates with various display configurations.
package display

import (
	"os"
	"time"

	"github.com/fatih/color"
)

// ColorScheme defines colors for different latency ranges
type ColorScheme struct {
	Excellent *color.Color
	Good      *color.Color
	Degraded  *color.Color
	Poor      *color.Color
	Offline   *color.Color
	Header    *color.Color
	Progress  *color.Color
	Success   *color.Color
	Warning   *color.Color
}

// DefaultColorScheme returns the default color scheme
func DefaultColorScheme() *ColorScheme {
	return &ColorScheme{
		Excellent: color.New(color.FgGreen, color.Bold),
		Good:      color.New(color.FgYellow),
		Degraded:  color.New(color.FgHiYellow),
		Poor:      color.New(color.FgRed, color.Bold),
		Offline:   color.New(color.FgHiBlack),
		Header:    color.New(color.FgCyan, color.Bold),
		Progress:  color.New(color.FgBlue),
		Success:   color.New(color.FgGreen, color.Bold),
		Warning:   color.New(color.FgYellow, color.Bold),
	}
}

// NoColorScheme returns a color scheme that doesn't use colors
func NoColorScheme() *ColorScheme {
	// Create a color that doesn't apply any formatting
	noColor := color.New()
	return &ColorScheme{
		Excellent: noColor,
		Good:      noColor,
		Degraded:  noColor,
		Poor:      noColor,
		Offline:   noColor,
		Header:    noColor,
		Progress:  noColor,
		Success:   noColor,
		Warning:   noColor,
	}
}

// GetColorForLatency returns the appropriate color based on latency
func (cs *ColorScheme) GetColorForLatency(latency time.Duration) *color.Color {
	if latency == 0 {
		return cs.Offline
	}

	// Convert to milliseconds for comparison
	ms := float64(latency.Nanoseconds()) / 1000000

	switch {
	case ms < 10:
		return cs.Excellent
	case ms < 100:
		return cs.Good
	case ms < 500:
		return cs.Degraded
	default:
		return cs.Poor
	}
}

// GetColorForPacketLoss returns the appropriate color based on packet loss
func (cs *ColorScheme) GetColorForPacketLoss(loss float64) *color.Color {
	switch {
	case loss == 0:
		return cs.Success
	case loss < 5:
		return cs.Excellent
	case loss < 20:
		return cs.Warning
	default:
		return cs.Poor
	}
}

// GetHostStatusIcon returns the appropriate icon for host status
func GetHostStatusIcon(status HostStatus) string {
	switch status {
	case StatusOnline:
		return "●"
	case StatusOffline:
		return "○"
	default:
		return "?"
	}
}

// IsTerminalColorSupported checks if the terminal supports colors
func IsTerminalColorSupported() bool {
	// Check NO_COLOR environment variable (standard convention)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if stdout is a terminal
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
