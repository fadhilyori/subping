// Package display provides terminal output formatting and progress tracking
// for network scanning results. It supports colored output, sortable tables,
// and real-time progress updates with various display configurations.
package display

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/olekukonko/tablewriter"
)

// TerminalDisplay implements Display interface for terminal output
type TerminalDisplay struct {
	config      DisplayConfig
	colorScheme *ColorScheme
	progress    *ProgressTracker
	table       *tablewriter.Table
}

// NewTerminalDisplay creates a new terminal display instance
func NewTerminalDisplay(config DisplayConfig) *TerminalDisplay {
	td := &TerminalDisplay{
		config: config,
	}

	// Initialize color scheme
	if config.EnabledColors && IsTerminalColorSupported() {
		td.colorScheme = DefaultColorScheme()
	} else {
		td.colorScheme = NoColorScheme()
	}

	return td
}

func (td *TerminalDisplay) ShowHeader(network string, ipRange string, totalHosts int, workers int, count int, interval string, timeout string, version string) {
	if td.config.EnabledProgress {
		throttle := td.config.progressThrottle
		if throttle == 0 {
			throttle = 200 * time.Millisecond // Default
		}
		td.progress = NewProgressTracker(totalHosts, td.colorScheme, true, throttle)
	}

	// Simple list format - no boxes, no width constraints
	td.colorScheme.Header.Printf("Subping %s\n", version)
	td.colorScheme.Header.Printf("Network: %s\n", network)
	td.colorScheme.Header.Printf("Range: %s\n", ipRange)
	td.colorScheme.Header.Printf("Hosts: %d | Workers: %d | Packets: %d\n", totalHosts, workers, count)
	fmt.Println()
}

// UpdateProgress updates the progress display
func (td *TerminalDisplay) UpdateProgress(current, total int, currentIP string, onlineCount int) {
	if td.progress != nil {
		td.progress.Update(current, currentIP, onlineCount)
	}

	// Call the user callback if provided
	if td.config.ProgressCallback != nil {
		td.config.ProgressCallback(current, total, currentIP, onlineCount)
	}
}

func (td *TerminalDisplay) ShowResults(results []HostResult) {
	if td.progress != nil {
		td.progress.Finish()
	}

	td.sortResults(results)

	td.table = tablewriter.NewWriter(os.Stdout)
	td.table.Header("IP Address", "Status", "Latency (Min/Avg/Max)", "Loss %", "Jitter")

	for _, result := range results {
		row := td.formatResultRow(result)
		td.table.Append(row)
	}

	td.table.Render()
}

func (td *TerminalDisplay) ShowSummary(totalHosts int, onlineHosts int, offlineHosts int, executionTime time.Duration) {
	var healthScore int
	if td.config.UseEnhancedHealthScore {
		// For enhanced scoring, we need the actual results
		// Since we don't have them here, fall back to basic scoring
		healthScore = td.calculateNetworkHealthScore(onlineHosts, totalHosts)
	} else {
		healthScore = td.calculateNetworkHealthScore(onlineHosts, totalHosts)
	}

	fmt.Println()

	summaryColor := td.colorScheme.Success
	if float64(onlineHosts)/float64(totalHosts) < 0.5 {
		summaryColor = td.colorScheme.Warning
	}

	summaryColor.Printf("Network Health Score: %d/100 (%s)\n",
		healthScore,
		td.getHealthDescription(healthScore))

	fmt.Printf("Scan completed in %s | ", executionTime.Round(time.Millisecond))
	summaryColor.Printf("%d hosts online", onlineHosts)
	fmt.Printf(" | %d hosts offline\n", offlineHosts)
}

func (td *TerminalDisplay) sortResults(results []HostResult) {
	switch td.config.SortBy {
	case SortByLatency:
		sort.Slice(results, func(i, j int) bool {
			return results[i].AvgRtt < results[j].AvgRtt
		})
	case SortByLoss:
		sort.Slice(results, func(i, j int) bool {
			return results[i].PacketLoss < results[j].PacketLoss
		})
	case SortByJitter:
		sort.Slice(results, func(i, j int) bool {
			return results[i].Jitter < results[j].Jitter
		})
	default: // SortByIP
		td.sortResultsByIP(results)
	}
}

// sortResultsByIP sorts results by IP address using cached byte representations
func (td *TerminalDisplay) sortResultsByIP(results []HostResult) {
	// Create helper struct with cached IP bytes for efficient comparison
	type sortHelper struct {
		result  HostResult
		ipBytes []byte
	}

	helpers := make([]sortHelper, len(results))
	for i, result := range results {
		helpers[i] = sortHelper{
			result:  result,
			ipBytes: result.IP.To16(),
		}
	}

	// Sort using cached bytes
	sort.Slice(helpers, func(i, j int) bool {
		return bytes.Compare(helpers[i].ipBytes, helpers[j].ipBytes) < 0
	})

	// Copy sorted results back to original slice
	for i, helper := range helpers {
		results[i] = helper.result
	}
}

func (td *TerminalDisplay) formatResultRow(result HostResult) []string {
	ipStr := result.IP.String()
	statusIcon := GetHostStatusIcon(result.Status)

	var latencyStr string
	if result.Status == StatusOnline {
		latencyStr = fmt.Sprintf("%s / %s / %s",
			td.formatDuration(result.MinRtt),
			td.formatDuration(result.AvgRtt),
			td.formatDuration(result.MaxRtt))
	} else {
		latencyStr = "—"
	}

	var lossStr string
	if result.Status == StatusOnline {
		lossStr = fmt.Sprintf("%.2f", result.PacketLoss)
	} else {
		lossStr = "100.00"
	}

	var jitterStr string
	if result.Status == StatusOnline {
		jitterStr = td.formatDuration(result.Jitter)
	} else {
		jitterStr = "—"
	}

	if td.config.EnabledColors {
		rowColor := td.colorScheme.GetColorForLatency(result.AvgRtt)

		ipStr = rowColor.Sprint(ipStr)
		statusIcon = rowColor.Sprint(statusIcon)
		if result.Status == StatusOnline {
			latencyStr = rowColor.Sprint(latencyStr)
			lossStr = rowColor.Sprint(lossStr)
			jitterStr = rowColor.Sprint(jitterStr)
		}
	}

	return []string{ipStr, statusIcon, latencyStr, lossStr, jitterStr}
}

func (td *TerminalDisplay) formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	if d < time.Second {
		ms := float64(d.Nanoseconds()) / 1000000
		return fmt.Sprintf("%.1fms", ms)
	}

	return d.String()
}

// calculateNetworkHealthScore calculates a basic network health score (0-100)
func (td *TerminalDisplay) calculateNetworkHealthScore(onlineHosts, totalHosts int) int {
	if totalHosts == 0 {
		return 0
	}

	// Basic calculation based on percentage of online hosts
	onlinePercentage := float64(onlineHosts) / float64(totalHosts)
	score := int(onlinePercentage * 100)

	return score
}

// calculateEnhancedNetworkHealthScore calculates an enhanced network health score (0-100)
// incorporating latency, packet loss, and jitter metrics beyond just online/offline ratio
func (td *TerminalDisplay) calculateEnhancedNetworkHealthScore(results []HostResult) int {
	if len(results) == 0 {
		return 0
	}

	var totalScore float64
	var validHosts int

	for _, result := range results {
		if result.Status == StatusOnline {
			validHosts++
			hostScore := 100.0

			// Deduct for high latency (>100ms)
			if result.AvgRtt > 100*time.Millisecond {
				hostScore -= 20
			} else if result.AvgRtt > 50*time.Millisecond {
				hostScore -= 10
			}

			// Deduct for packet loss
			hostScore -= result.PacketLoss * 0.5

			// Deduct for high jitter (>50ms)
			if result.Jitter > 50*time.Millisecond {
				hostScore -= 15
			}

			totalScore += math.Max(0, hostScore)
		}
	}

	if validHosts == 0 {
		return 0
	}

	// Weight by online percentage
	onlinePercentage := float64(validHosts) / float64(len(results))
	return int(totalScore / float64(validHosts) * onlinePercentage)
}

// getHealthDescription returns a description for the health score
func (td *TerminalDisplay) getHealthDescription(score int) string {
	switch {
	case score >= 80:
		return "Excellent"
	case score >= 60:
		return "Good"
	case score >= 40:
		return "Fair"
	case score >= 20:
		return "Poor"
	default:
		return "Critical"
	}
}
