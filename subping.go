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
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/fadhilyori/subping/internal/ping"
	"github.com/fadhilyori/subping/pkg/network"
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
	Results map[string]ping.Result

	// TotalResults represents the total number of ping results collected.
	TotalResults int

	// MaxWorkers specifies the maximum number of concurrent workers to use.
	MaxWorkers int

	// pinger is the ping implementation (real or mock)
	pinger ping.Pinger

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


// NewSubping creates a new Subping instance with the provided options.
func NewSubping(opts *Options) (*Subping, error) {
	if opts.Subnet == "" {
		return nil, errors.New("subnet should be in CIDR notation and cannot be empty")
	}

	if opts.Count < 1 {
		return nil, errors.New("count should be more than zero (0)")
	}

	if opts.MaxWorkers < 1 {
		return nil, errors.New("max workers should be more than zero (0)")
	}

	if opts.Timeout < 0 {
		return nil, errors.New("timeout cannot be negative")
	}

	if opts.Interval < 0 {
		return nil, errors.New("interval cannot be negative")
	}

	ips, err := network.NewSubnetHostsIteratorFromCIDRString(opts.Subnet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subnet: %w", err)
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
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	instance := &Subping{
		TargetsIterator: ips,
		Count:           opts.Count,
		Interval:        opts.Interval,
		Timeout:         opts.Timeout,
		BatchSize:       int64(batchLimit),
		MaxWorkers:      opts.MaxWorkers,
		pinger:          ping.NewPinger(), // Auto-detect based on environment
		logger:          logrus.New(),
	}

	instance.logger.SetLevel(logLevel)

	return instance, nil
}

// NewSubpingWithPinger creates a new Subping instance with a custom pinger implementation
// This allows dependency injection for testing or special use cases
func NewSubpingWithPinger(opts *Options, pinger ping.Pinger) (*Subping, error) {
	if opts.Subnet == "" {
		return nil, errors.New("subnet should be in CIDR notation and cannot be empty")
	}

	if opts.Count < 1 {
		return nil, errors.New("count should be more than zero (0)")
	}

	if opts.MaxWorkers < 1 {
		return nil, errors.New("max workers should be more than zero (0)")
	}

	if opts.Timeout < 0 {
		return nil, errors.New("timeout cannot be negative")
	}

	if opts.Interval < 0 {
		return nil, errors.New("interval cannot be negative")
	}

	ips, err := network.NewSubnetHostsIteratorFromCIDRString(opts.Subnet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subnet: %w", err)
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
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	instance := &Subping{
		TargetsIterator: ips,
		Count:           opts.Count,
		Interval:        opts.Interval,
		Timeout:         opts.Timeout,
		BatchSize:       int64(batchLimit),
		MaxWorkers:      opts.MaxWorkers,
		pinger:          pinger, // Use the provided pinger
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
	s.Results = make(map[string]ping.Result)

	syncMap.Range(func(key, value any) bool {
		s.Results[key.(string)] = value.(ping.Result)

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

		p, err := s.pinger.Ping(target, s.Count, s.Interval, s.Timeout)
		if err != nil {
			s.logger.WithField("worker", id).Debugf("Ping failed for %s: %v", target, err)
			// Store empty result for failed pings
			p = ping.Result{}
		}

		sm.Store(target, p)

		time.Sleep(s.Interval)
	}
}

// GetOnlineHosts returns a map of online hosts and their corresponding ping results,
// as well as the total number of online hosts.
func (s *Subping) GetOnlineHosts() (map[string]ping.Result, int) {
	r := make(map[string]ping.Result)

	for ip, stats := range s.Results {
		if stats.PacketsRecv > 0 {
			r[ip] = stats
		}
	}

	return r, len(r)
}

// RunPing performs a ping operation to the specified IP address.
// It sends the specified number of ping requests with the given interval and timeout.
// This function delegates to the internal ping package for implementation.
func RunPing(ipAddress string, count int, interval time.Duration, timeout time.Duration) ping.Statistics {
	return ping.RunPing(ipAddress, count, interval, timeout)
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
