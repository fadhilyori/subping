package subping

import (
	"errors"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"github.com/go-ping/ping"
)

type Subping struct {
	targets []net.IP
	count   int
	timeout time.Duration
	numJobs int
	results map[string]*ping.Statistics
}

type Options struct {
	Targets []net.IP
	Count   int
	Timeout time.Duration
	NumJobs int
}

func NewSubping(opts *Options) (Subping, error) {
	if len(opts.Targets) < 1 {
		return Subping{}, errors.New("target cannot empty")
	}

	if opts.Count < 1 {
		return Subping{}, errors.New("count should be more than zero (0)")
	}

	if opts.NumJobs < 1 {
		return Subping{}, errors.New("number of jobs should be more than zero (0)")
	}

	return Subping{
		targets: opts.Targets,
		count:   opts.Count,
		timeout: opts.Timeout,
		numJobs: opts.NumJobs,
	}, nil
}

func (s *Subping) Run() {
	var (
		wg                  sync.WaitGroup
		resultsFromRoutines []map[string]*ping.Statistics
	)
	splitedIPList := partitionSlice(s.targets, s.numJobs)

	startJob := func(targets []net.IP) {
		defer wg.Done()

		result := make(map[string]*ping.Statistics)

		for _, target := range targets {
			p := RunPing(target, s.count, s.timeout)
			result[target.String()] = p
		}

		resultsFromRoutines = append(resultsFromRoutines, result)
	}

	for _, job := range splitedIPList {
		wg.Add(1)
		go startJob(job)
	}

	wg.Wait()

	s.results = func(s []map[string]*ping.Statistics) map[string]*ping.Statistics {
		flattened := make(map[string]*ping.Statistics)

		// Flatten the slice into a map
		for _, m := range s {
			for k, v := range m {
				flattened[k] = v
			}
		}

		return flattened
	}(resultsFromRoutines)
}

func (s *Subping) GetResults() map[string]*ping.Statistics {
	return s.results
}

func (s *Subping) GetOnlineHosts() map[string]*ping.Statistics {
	r := make(map[string]*ping.Statistics)

	for ip, stats := range s.results {
		if stats.PacketsRecv > 0 {
			r[ip] = stats
		}
	}

	return r
}

func RunPing(ipAddress net.IP, count int, timeout time.Duration) *ping.Statistics {
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
