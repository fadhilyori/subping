package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sort"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/fadhilyori/subping"
	"github.com/spf13/cobra"
)

var (
	pingCount           int
	pingTimeoutStr      string
	pingIntervalStr     string
	pingMaxWorkers      int
	subpingVersion      = "dev"
	showOfflineHostList bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "subping [flags] [network subnet]",
		Version: subpingVersion,
		Short:   "A tool for pinging IP addresses in a subnet",
		Long:    "Subping is a command-line tool that allows you to ping IP addresses within a specified subnet range.",
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run:     runSubping,
		PreRun: func(cmd *cobra.Command, args []string) {
			figure.NewFigure("subping", "larry3d", true).Print()
			fmt.Println(cmd.Version)
			fmt.Print("\n\n")
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
	flags.StringVarP(&pingTimeoutStr, "timeout", "t", "80ms",
		"Specifies the maximum ping timeout duration for each ping request.",
	)
	flags.StringVarP(&pingIntervalStr, "interval", "i", "300ms",
		"Specifies the time duration between each ping request.",
	)
	flags.BoolVar(&showOfflineHostList, "offline", false,
		"Specify whether to display the list of offline hosts.",
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runSubping(_ *cobra.Command, args []string) {
	subnetString := args[0]

	startTime := time.Now()

	pingTimeout, err := time.ParseDuration(pingTimeoutStr)
	if err != nil {
		log.Fatal(err.Error())
	}

	pingInterval, err := time.ParseDuration(pingIntervalStr)
	if err != nil {
		log.Fatal(err.Error())
	}

	s, err := subping.NewSubping(&subping.Options{
		Subnet:     subnetString,
		Count:      pingCount,
		Interval:   pingInterval,
		Timeout:    pingTimeout * time.Duration(pingCount),
		MaxWorkers: pingMaxWorkers,
		LogLevel:   "error",
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("Network        : %s\n", s.TargetsIterator.IPNet.String())
	fmt.Printf("IP Ranges      : %s - %s\n",
		s.TargetsIterator.FirstIP.String(), s.TargetsIterator.LastIP.String(),
	)
	fmt.Printf("Total hosts    : %d\n", s.TargetsIterator.TotalHosts)
	fmt.Printf("Total workers  : %d\n", s.MaxWorkers)
	fmt.Printf("Count          : %d\n", s.Count)
	fmt.Printf("Interval       : %s\n", s.Interval.String())
	fmt.Printf("Timeout        : %s\n", pingTimeoutStr)
	fmt.Println(`-------------------------------------------------------------------------------`)
	fmt.Printf("| %-39s | %-16s | %-14s |\n", "IP Address", "Avg Latency", "Packet Loss")
	fmt.Println(`-------------------------------------------------------------------------------`)

	s.Run()

	results, totalHostOnline := s.GetOnlineHosts()

	// Extract keys into a slice
	keys := make([]net.IP, 0, len(results))
	for key := range results {
		keys = append(keys, net.ParseIP(key))
	}

	// Sort the keys Based on byte comparison
	sort.Slice(keys, func(i, j int) bool {
		return bytes.Compare(keys[i].To16(), keys[j].To16()) < 0
	})

	for _, ip := range keys {
		// convert bytes to string in each line of IP
		ipString := ip.String()
		stats := results[ipString]
		packetLossPercentageStr := fmt.Sprintf("%.2f %%", stats.PacketLoss)

		fmt.Printf(
			"| %-39s | %-16s | %-14s |\n",
			ipString, stats.AvgRtt.String(), packetLossPercentageStr)
	}

	fmt.Println(`-------------------------------------------------------------------------------`)

	if showOfflineHostList {
		fmt.Println("\nOffline hosts :")
		for ip, stats := range s.Results {
			if stats.PacketsRecv == 0 {
				fmt.Printf(
					" - %s\t(Loss: %s, Latency: %s)\n",
					ip, fmt.Sprintf("%.2f %%", stats.PacketLoss), stats.AvgRtt.String(),
				)
			}
		}
	}

	elapsed := time.Since(startTime)
	totalHostOffline := s.TargetsIterator.TotalHosts - totalHostOnline

	fmt.Printf("\nTotal Hosts Online  : %d\n", totalHostOnline)
	fmt.Printf("Total Hosts Offline : %d\n", totalHostOffline)
	fmt.Printf("Execution time      : %s\n\n", elapsed.String())
}
