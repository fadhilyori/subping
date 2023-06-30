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

	"github.com/go-ping/ping"
	_ "go.uber.org/automaxprocs"
)

var (
	wg sync.WaitGroup
)

func doPing(ipAddress string, count int, timeout time.Duration) *ping.Statistics {
	pinger, err := ping.NewPinger(ipAddress)
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

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func partitionStrings(arr []string, numPartitions int) [][]string {
	arrSize := len(arr)
	chunkSize := int(math.Ceil(float64(arrSize) / float64(numPartitions)))

	var result [][]string

	for i := 0; i < arrSize; i += chunkSize {
		end := i + chunkSize
		if end > arrSize {
			end = arrSize
		}

		result = append(result, arr[i:end])
	}

	return result
}

func getIPList(subnet string) []string {
	var ips []string
	ip, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		fmt.Printf("Failed to parse subnet: %v\n", err)
		return []string{}
	}

	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	return ips
}

func main() {
	subnetString := os.Args[1]
	ips := getIPList(subnetString)
	workersCount := runtime.NumCPU()
	ipsSplit := partitionStrings(ips, workersCount)

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
