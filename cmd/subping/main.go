package main

import (
	"fmt"
	"log"
	"runtime"
	"sort"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/fadhilyori/subping"
	"github.com/fadhilyori/subping/pkg/network"
	"github.com/spf13/cobra"
)

var (
	pingCount      int
	pingTimeoutStr string
	pingNumJobs    int
	subpingVersion string = "latest"
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

	flags.IntVarP(&pingCount, "count", "c", 3,
		"Specifies the number of ping attempts for each IP address.",
	)
	flags.IntVarP(&pingNumJobs,
		"job", "n", runtime.NumCPU(),
		"Specifies the number of maximum concurrent jobs spawned to perform ping operations."+
			"\nThe default value is equal to the number of CPUs available on the system.",
	)
	flags.StringVarP(&pingTimeoutStr, "timeout", "t", "300ms",
		"Specifies the maximum ping timeout duration. The default value is \"300ms\".",
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

	ips, err := network.GenerateIPListFromCIDRString(subnetString)
	if err != nil {
		log.Fatal(err.Error())
	}

	s, err := subping.NewSubping(&subping.Options{
		Targets: ips,
		Count:   pingCount,
		Timeout: pingTimeout,
		NumJobs: pingNumJobs,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("Network\t\t: %s\n", subnetString)
	fmt.Printf("IP Ranges\t: %v - %v\n", ips[0], ips[len(ips)-1])
	fmt.Printf("Total hosts\t: %d\n", len(ips))
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

	elapsed := time.Since(startTime)
	fmt.Printf("Execution time: %s\n", elapsed.String())
}
