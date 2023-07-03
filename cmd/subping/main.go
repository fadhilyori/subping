package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/fadhilyori/subping"
	"github.com/fadhilyori/subping/pkg/network"
)

type flagConfig struct {
	count   int
	timeout time.Duration
	numJobs int
}

func main() {
	subnetString := os.Args[len(os.Args)-1]

	config := flagConfig{}
	var timeoutStr string

	flag.IntVar(&config.count, "c", 3, "Specifies the number of ping attempts for each IP address.")
	flag.IntVar(&config.numJobs, "n", runtime.NumCPU(), "Specifies the number of maximum concurrent jobs spawned to perform ping operations.\nThe default value is equal to the number of CPUs available on the system.")
	flag.StringVar(&timeoutStr, "t", "300ms", "Specifies the maximum ping timeout duration. The default value is \"300ms\".")
	flag.Usage = func() {
		_, err := fmt.Fprintln(flag.CommandLine.Output(), "Usage:\n\nsubping [OPTIONS] <network subnet>\n\nOptions:")
		if err != nil {
			return
		}
		flag.PrintDefaults()
	}
	flag.Parse()

	startTime := time.Now()

	t, err := time.ParseDuration(timeoutStr)
	if err != nil {
		log.Fatal(err.Error())
	}

	config.timeout = t

	ips, err := network.GenerateIPListFromCIDRString(subnetString)
	if err != nil {
		log.Fatal(err.Error())
	}

	s, err := subping.NewSubping(&subping.Options{
		Targets: ips,
		Count:   config.count,
		Timeout: config.timeout,
		NumJobs: config.numJobs,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("Network\t\t: %s\n", subnetString)
	fmt.Printf("IP Ranges\t: %v - %v\n", ips[0], ips[len(ips)-1])
	fmt.Printf("Total hosts\t: %d\n", len(ips))
	fmt.Println("---------------------------------------")
	fmt.Println("| IP Address       | Avg Latency      |")
	fmt.Println("---------------------------------------")
	fmt.Printf("Pinging...")

	s.Run()

	results := s.GetOnlineHosts()

	// Extract keys into a slice
	keys := make([]net.IP, 0, len(results))
	for key := range results {
		keys = append(keys, net.ParseIP(key))
	}

	// Sort the keys Based on byte comparison
	sort.Slice(keys, func(i, j int) bool {
		return bytes.Compare(keys[i].To4(), keys[j].To4()) < 0
	})

	fmt.Print("\r")

	for _, ip := range keys {
		// convert bytes to string in each line of IP
		ipString := ip.String()
		stats := results[ipString]

		fmt.Printf("| %-15s | %-15s |\n", ipString, stats.AvgRtt.String())
	}

	fmt.Println("---------------------------------------")

	elapsed := time.Since(startTime)
	fmt.Printf("Execution time: %s\n", elapsed.String())
}
