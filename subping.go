// Package subping provides a utility for concurrently pinging multiple IP addresses and collecting the results.
//
// The package includes functionality for running ping operations on multiple IP addresses concurrently,
// calculating ping statistics, and partitioning data for parallel processing.
//
// Example usage:
//
//	// Create options for Subping
//	opts := &subping.Options{
//	    LogLevel:   "info",
//	    Subnet:     "192.168.0.0/24",
//	    Count:      5,
//	    Interval:   time.Second,
//	    Timeout:    2 * time.Second,
//	    MaxWorkers: 10,
//	}
//
//	// Create a new Subping instance
//	sp, err := subping.NewSubping(opts)
//	if err != nil {
//	    log.Fatalf("Failed to create Subping instance: %v", err)
//	}
//
//	// Run the Subping process
//	sp.Run()
//
//	// Get the online hosts and their statistics
//	onlineHosts, total := sp.GetOnlineHosts()
//	fmt.Printf("Online Hosts: %v\n", onlineHosts)
//	fmt.Printf("Total Online Hosts: %d\n", total)
package subping

import (
	"errors"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/fadhilyori/subping/pkg/network"
	ping "github.com/prometheus-community/pro-bing"
)

// Subping is a utility for concurrently pinging multiple IP addresses and collecting the results.
type Subping struct {
	// TargetsIterator is an iterator for the target IP addresses to ping.
	TargetsIterator *network.SubnetHostsIterator

	// Count is the number of ping requests to send for each target.
	Count int

	// Interval is the time duration between each ping request.
	Interval time.Duration

	// Timeout specifies the timeout duration before exiting each target.
	Timeout time.Duration

	// BatchSize is the number of concurrent ping jobs to execute.
	BatchSize int64

	// Results stores the ping results for each target IP address.
	Results map[string]Result

	// TotalResults represents the total number of ping results collected.
	TotalResults int

	// MaxWorkers specifies the maximum number of concurrent workers to use.
	MaxWorkers int

	logger *logrus.Logger
}

// Options holds the configuration options for creating a new Subping instance.
type Options struct {
	// LogLevel sets the log levels for the Subping instance.
	LogLevel string

	// Subnet is the subnet to scan for IP addresses to ping.
	Subnet string

	// Count is the number of ping requests to send for each target.
	Count int

	// Interval is the time duration between each ping request.
	Interval time.Duration

	// Timeout specifies the timeout duration before exiting each target.
	Timeout time.Duration

	// MaxWorkers specifies the maximum number of concurrent workers to use.
	MaxWorkers int
}

// Result contains the statistics and metrics for a single ping operation.
type Result struct {
	// AvgRtt is the average round-trip time of the ping requests.
	AvgRtt time.Duration

	// PacketLoss is the percentage of packets lost during the ping operation.
	PacketLoss float64

	// PacketsSent is the number of packets sent for the ping operation.
	PacketsSent int

	// PacketsRecv is the number of packets received for the ping operation.
	PacketsRecv int

	// PacketsRecvDuplicates is the number of duplicate packets received.
	PacketsRecvDuplicates int
}

// NewSubping creates a new Subping instance with the provided options.
func NewSubping(opts *Options) (*Subping, error) {
	if opts.Subnet == "" {
		return nil, errors.New("subnet should be in CIDR notation and cannot empty")
	}

	if opts.Count < 1 {
		return nil, errors.New("count should be more than zero (0)")
	}

	if opts.MaxWorkers < 1 {
		return nil, errors.New("max workers should be more than zero (0)")
	}

	ips, err := network.NewSubnetHostsIteratorFromCIDRString(opts.Subnet)
	if err != nil {
		log.Fatal(err.Error())
	}

	batchLimit, err := calculateMaxPartitionSize(ips.TotalHosts, opts.MaxWorkers)
	if err != nil {
		return nil, err
	}

	if opts.LogLevel == "" {
		opts.LogLevel = "error"
	}

	logLevel, err := logrus.ParseLevel(opts.LogLevel)
	if err != nil {
		return nil, errors.New("max workers should be more than zero (0)")
	}

	instance := &Subping{
		TargetsIterator: ips,
		Count:           opts.Count,
		Interval:        opts.Interval,
		Timeout:         opts.Timeout,
		BatchSize:       int64(batchLimit),
		MaxWorkers:      opts.MaxWorkers,
		logger:          logrus.New(),
	}

	instance.logger.SetLevel(logLevel)

	return instance, nil
}

// Run starts the Subping process, concurrently pinging the target IP addresses.
// It spawns worker goroutines, assigns tasks to them, waits for them to finish,
// and collects the results.
func (s *Subping) Run() {
	var (
		// syncMap to store the results from workers.
		syncMap sync.Map

		// wg WaitGroup to synchronize the workers.
		wg sync.WaitGroup

		// jobChannel to distribute tasks to workers.
		jobChannel = make(chan string, s.BatchSize)
	)

	// Spawn the worker goroutines.
	for i := int64(0); i < int64(s.MaxWorkers); i++ {
		wg.Add(1)
		go s.startWorker(i, &wg, &syncMap, jobChannel)
	}

	s.logger.Debugf("Spawned %d workers.\n", s.MaxWorkers)

	s.logger.Debugln("Assigning task to all workers.")
	for ip := s.TargetsIterator.Next(); ip != nil; ip = s.TargetsIterator.Next() {
		ipString := ip.String()
		jobChannel <- ipString
		s.logger.Tracef("Assigned task: %s\n", ipString)
	}

	s.logger.Debugln("Waiting all workers finish their jobs.")
	close(jobChannel)
	wg.Wait()

	s.logger.Debugln("All workers already stopped. Storing the results.")
	s.Results = make(map[string]Result)

	syncMap.Range(func(key, value any) bool {
		s.Results[key.(string)] = value.(Result)

		return true
	})
	s.TotalResults = len(s.Results)
	s.logger.Debugln("Run finished. All task done..")
}

// startWorker is a worker goroutine that performs the ping task assigned to it.
// It collects the ping results and stores them in the sync.Map.
func (s *Subping) startWorker(id int64, wg *sync.WaitGroup, sm *sync.Map, c <-chan string) {
	defer wg.Done()

	for target := range c {
		s.logger.WithField("worker", id).Tracef("Got task %s.\n", target)

		p := RunPing(target, s.Count, s.Interval, s.Timeout)
		sm.Store(target, Result{
			AvgRtt:                p.AvgRtt,
			PacketLoss:            p.PacketLoss,
			PacketsSent:           p.PacketsSent,
			PacketsRecv:           p.PacketsRecv,
			PacketsRecvDuplicates: p.PacketsRecvDuplicates,
		})

		time.Sleep(s.Interval)
	}
}

// GetOnlineHosts returns a map of online hosts and their corresponding ping results,
// as well as the total number of online hosts.
func (s *Subping) GetOnlineHosts() (map[string]Result, int) {
	r := make(map[string]Result)

	for ip, stats := range s.Results {
		if stats.PacketsRecv > 0 {
			r[ip] = stats
		}
	}

	return r, len(r)
}

// RunPing performs a ping operation to the specified IP address.
// It sends the specified number of ping requests with the given interval and timeout.
func RunPing(ipAddress string, count int, interval time.Duration, timeout time.Duration) ping.Statistics {
	pinger, err := ping.NewPinger(ipAddress)
	if err != nil {
		logrus.Printf("Failed to create pinger for IP Address: %s\n", ipAddress)
		return ping.Statistics{}
	}

	pinger.Count = count
	pinger.Interval = interval

	if timeout > 0 {
		pinger.Timeout = timeout
	}

	if runtime.GOOS == "windows" {
		pinger.SetPrivileged(true)
	}

	err = pinger.Run()
	if err != nil {
		logrus.Printf("Failed to ping the address %s, %v\n", ipAddress, err.Error())
		return ping.Statistics{}
	}

	return *pinger.Statistics()
}

// calculateMaxPartitionSize calculates the maximum size of each partition given the total data size and the desired number of partitions.
func calculateMaxPartitionSize(dataSize int, numPartitions int) (int, error) {
	maxPartitionSize := dataSize / numPartitions
	remainder := dataSize % numPartitions

	if remainder != 0 {
		maxPartitionSize++
	}

	if maxPartitionSize < 0 {
		return 0, errors.New("the value exceeds the range of int")
	}

	return maxPartitionSize, nil
}
