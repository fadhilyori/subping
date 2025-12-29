package main

import (
	"fmt"
	"log"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/fadhilyori/subping"
	"github.com/fadhilyori/subping/internal/display"
	"github.com/spf13/cobra"
)

var (
	pingCount           int
	pingTimeoutStr      string
	pingIntervalStr     string
	pingMaxWorkers      int
	subpingVersion      = "dev"
	showOfflineHostList bool
	sortBy              string
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "subping [flags] [network subnet]",
		Version: subpingVersion,
		Short:   "A tool for pinging IP addresses in a subnet",
		Long:    "Subping is a command-line tool that allows you to ping IP addresses within a specified subnet range.",
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			runSubping(cmd, args)
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			figure.NewFigure("subping", "larry3d", true).Print()
			fmt.Print("\n")
		},
	}

	flags := rootCmd.Flags()

	flags.IntVarP(&pingCount, "count", "c", 1,
		"Specifies the number of ping attempts for each IP address.",
	)
	flags.IntVarP(&pingMaxWorkers,
		"job", "n", 128,
		"Specifies the number of maximum concurrent jobs spawned to perform ping operations.",
	)
	flags.StringVarP(&pingTimeoutStr, "timeout", "t", "1s",
		"Specifies the maximum ping timeout duration for each ping request.",
	)
	flags.StringVarP(&pingIntervalStr, "interval", "i", "300ms",
		"Specifies the time duration between each ping request.",
	)
	flags.BoolVar(&showOfflineHostList, "offline", false,
		"Specify whether to display the list of offline hosts.",
	)
	flags.StringVar(&sortBy, "sort", "ip",
		"Sort results by: ip, latency, loss, jitter (default: ip)",
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runSubping(rootCmd *cobra.Command, args []string) {
	subnetString := args[0]

	startTime := time.Now()

	pingTimeout, err := time.ParseDuration(pingTimeoutStr)
	if err != nil {
		log.Fatalf("Invalid timeout format '%s': %v\nValid examples: 1s, 500ms, 1m30s", pingTimeoutStr, err)
	}

	pingInterval, err := time.ParseDuration(pingIntervalStr)
	if err != nil {
		log.Fatalf("Invalid interval format '%s': %v\nValid examples: 300ms, 1s, 2s", pingIntervalStr, err)
	}

	sortOption := display.SortByIP
	switch sortBy {
	case "latency":
		sortOption = display.SortByLatency
	case "loss":
		sortOption = display.SortByLoss
	case "jitter":
		sortOption = display.SortByJitter
	case "ip":
		sortOption = display.SortByIP
	default:
		log.Fatalf("Invalid sort option '%s'. Valid options: ip, latency, loss, jitter", sortBy)
	}

	displayConfig := display.DisplayConfig{
		EnabledColors:          true,
		EnabledProgress:        true,
		SortBy:                 sortOption,
		UseEnhancedHealthScore: true,
	}

	d := display.NewDisplay(displayConfig)

	s, err := subping.NewSubping(&subping.Options{
		Subnet:     subnetString,
		Count:      pingCount,
		Interval:   pingInterval,
		Timeout:    pingTimeout,
		MaxWorkers: pingMaxWorkers,
		LogLevel:   "error",
		ProgressCallback: func(current, total int, currentIP string, onlineCount int) {
			d.UpdateProgress(current, total, currentIP, onlineCount)
		},
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	d.ShowHeader(
		s.TargetsIterator.IPNet.String(),
		fmt.Sprintf("%s - %s", s.TargetsIterator.FirstIP.String(), s.TargetsIterator.LastIP.String()),
		s.TargetsIterator.TotalHosts,
		s.MaxWorkers,
		s.Count,
		s.Interval.String(),
		pingTimeoutStr,
		rootCmd.Version,
	)

	s.Run()

	onlineResults, totalHostOnline := s.GetOnlineHosts()

	// Calculate proper capacity based on what we'll actually store
	var estimatedCapacity int
	if showOfflineHostList {
		estimatedCapacity = s.TargetsIterator.TotalHosts // All hosts (online + offline)
	} else {
		estimatedCapacity = totalHostOnline // Only online hosts
	}

	allResults := make([]display.HostResult, 0, estimatedCapacity)

	for ip, result := range onlineResults {
		allResults = append(allResults, display.ConvertPingResult(ip, result))
	}

	if showOfflineHostList {
		for ip, result := range s.Results {
			if result.PacketsRecv == 0 {
				allResults = append(allResults, display.ConvertPingResult(ip, result))
			}
		}
	}

	d.ShowResults(allResults)

	elapsed := time.Since(startTime)
	totalHostOffline := s.TargetsIterator.TotalHosts - totalHostOnline

	var scanRate float64
	if elapsed.Seconds() > 0 {
		scanRate = float64(s.TargetsIterator.TotalHosts) / elapsed.Seconds()
	}

	d.ShowSummary(s.TargetsIterator.TotalHosts, totalHostOnline, totalHostOffline, elapsed, scanRate)
}

// Avoid to ping with 0.0.0.0/0
