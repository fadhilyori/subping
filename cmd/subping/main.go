package main

import (
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/fadhilyori/subping/pkg/network"
	"github.com/go-ping/ping"
	_ "go.uber.org/automaxprocs"
)

var (
	wg sync.WaitGroup
)

func doPing(ipAddress net.IP, count int, timeout time.Duration) *ping.Statistics {
	pinger, err := ping.NewPinger(ipAddress.String())
	if err != nil {
		log.Printf("Failed to create pinger for IP Address: %s\n", ipAddress)
		return nil
	}

	pinger.Count = count
	pinger.Timeout = timeout
	err = pinger.Run()
	if err != nil {
		return nil
	}

	return pinger.Statistics()
}

func partitionSlice(arr []net.IP, numPartitions int) [][]net.IP {
	arrSize := len(arr)
	chunkSize := int(math.Ceil(float64(arrSize) / float64(numPartitions)))

	var result [][]net.IP

	for i := 0; i < arrSize; i += chunkSize {
		end := i + chunkSize
		if end > arrSize {
			end = arrSize
		}

		result = append(result, arr[i:end])
	}

	return result
}

func main() {
	subnetString := os.Args[1]
	ips, err := network.GenerateIPListFromCIDRString(subnetString)
	if err != nil {
		log.Fatal(err.Error())
	}
	workersCount := runtime.NumCPU()
	ipsSplit := partitionSlice(ips, workersCount)

	fmt.Printf("Network\t\t\t: %s\n", subnetString)
	fmt.Printf("IP addresses ranges\t: %v - %v\n", ips[0], ips[len(ips)-1])
	fmt.Printf("Total hosts\t\t: %d\n", len(ips))
	fmt.Printf("Partition\t\t: %d\n", len(ipsSplit))
	fmt.Println("-----------------------------------")
	fmt.Println("IP Address      | Avg Latency     |")
	fmt.Println("-----------------------------------")

	for _, ipJob := range ipsSplit {
		ipQueue := ipJob
		wg.Add(1)

		go func() {
			defer wg.Done()

			for _, ip := range ipQueue {

				p := doPing(ip, 3, 3*time.Millisecond)

				if p == nil {
					continue
				}

				if p.PacketsSent == 0 || p.PacketsRecv == 0 {
					continue
				}

				fmt.Printf("%-16s| %-16s|\n", ip, p.AvgRtt.String())
			}
		}()
	}

	wg.Wait()
	fmt.Println("-----------------------------------")
}
