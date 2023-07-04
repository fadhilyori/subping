package main

import (
	"fmt"
	"log"
	"net"
	"sort"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/fadhilyori/subping"
	"github.com/fadhilyori/subping/pkg/network"
	"github.com/spf13/cobra"
)

var (
	pingCount           int
	pingTimeoutStr      string
	pingIntervalStr     string
	pingNumJobs         int
	subpingVersion      = "latest"
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
			fmt.Print("\n\n")
		},
	}

	flags := rootCmd.Flags()

	flags.IntVarP(&pingCount, "count", "c", 1,
		"Specifies the number of ping attempts for each IP address.",
	)
	flags.IntVarP(&pingNumJobs,
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

func runSubping(cmd *cobra.Command, args []string) {
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

	ips, err := network.GenerateIPListFromCIDRString(subnetString)
	if err != nil {
		log.Fatal(err.Error())
	}

	totalHost := len(ips)

	s, err := subping.NewSubping(&subping.Options{
		Targets:  ips,
		Count:    pingCount,
		Interval: pingInterval,
		Timeout:  pingTimeout * time.Duration(pingCount),
		NumJobs:  pingNumJobs,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	_, cidr, err := net.ParseCIDR(subnetString)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("Network        : %s\n", cidr.String())
	fmt.Printf("IP Ranges      : %s - %s\n", ips[0].String(), ips[len(ips)-1].String())
	fmt.Printf("Total hosts    : %d\n", totalHost)
	fmt.Printf("Num of workers : %d\n", len(s.PartitionedTargets))
	fmt.Println(`---------------------------------------`)
	fmt.Println("| IP Address       | Avg Latency      |")
	fmt.Println(`---------------------------------------`)
	fmt.Printf("Pinging...")

	s.Run()

	results := s.GetOnlineHosts()

	// Extract keys into a slice
	keys := make([]string, 0, len(results))
	for key := range results {
		keys = append(keys, key)
	}

	// Sort the keys
	sort.Strings(keys)

	fmt.Print("\r")

	for _, ip := range keys {
		stats := results[ip]

		fmt.Printf("| %-16s | %-16s |\n", ip, stats.AvgRtt.String())
	}

	fmt.Println(`---------------------------------------`)

	if showOfflineHostList {
		fmt.Println("Offline hosts :")
		for ip, stats := range s.Results {
			if stats.PacketsRecv == 0 {
				fmt.Printf(" - %s\n", ip)
			}
		}
	}

	elapsed := time.Since(startTime)
	totalHostOnline := len(results)
	totalHostOffline := totalHost - totalHostOnline

	fmt.Printf("\nTotal Hosts Online  : %d\n", totalHostOnline)
	fmt.Printf("Total Hosts Offline : %d\n", totalHostOffline)
	fmt.Printf("Execution time      : %s\n\n", elapsed.String())
}
