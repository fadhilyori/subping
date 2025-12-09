package subping_test

import (
	"os"
	"sort"
	"testing"
	"time"

	"github.com/fadhilyori/subping"
	"github.com/fadhilyori/subping/pkg/network"
)

// TestMain sets up the test environment
// We force CI=true to ensure mock pinger is used for reliable testing
func TestMain(m *testing.M) {
	// Set CI environment variable to force mock pinger usage
	// This ensures tests run reliably without requiring network privileges
	os.Setenv("CI", "true")

	// Run tests
	code := m.Run()

	// Clean up
	os.Unsetenv("CI")

	os.Exit(code)
}

func TestRunSubping(t *testing.T) {
	type args struct {
		CIDR       string
		Count      int
		Timeout    time.Duration
		Interval   time.Duration
		MaxWorkers int
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantOnline  bool
		numOfOnline int
	}{
		{
			name: "Test with valid options",
			args: args{
				CIDR:       "127.0.0.1/31",
				Count:      1,
				Timeout:    300 * time.Millisecond,
				Interval:   300 * time.Millisecond,
				MaxWorkers: 1,
			},
			wantErr:    false,
			wantOnline: false,
			numOfOnline: 0,
		},
		{
			name: "Test with invalid Count",
			args: args{
				CIDR:       "127.0.0.0/29",
				Count:      -1,
				Timeout:    1 * time.Second,
				Interval:   300 * time.Millisecond,
				MaxWorkers: 2,
			},
			wantErr:    true,
			wantOnline: false,
			numOfOnline: 0,
		},
		{
			name: "Test with invalid MaxWorkers",
			args: args{
				CIDR:       "127.0.0.0/29",
				Count:      1,
				Timeout:    1 * time.Second,
				Interval:   300 * time.Millisecond,
				MaxWorkers: -2,
			},
			wantErr:    true,
			wantOnline: false,
			numOfOnline: 0,
		},
		{
			name: "Test with invalid negative Timeout",
			args: args{
				CIDR:       "127.0.0.0/29",
				Count:      1,
				Timeout:    -1 * time.Second,
				Interval:   300 * time.Millisecond,
				MaxWorkers: 2,
			},
			wantErr:    true,
			wantOnline: false,
			numOfOnline: 0,
		},
		{
			name: "Test with invalid negative Interval",
			args: args{
				CIDR:       "127.0.0.0/29",
				Count:      1,
				Timeout:    1 * time.Second,
				Interval:   -300 * time.Millisecond,
				MaxWorkers: 2,
			},
			wantErr:    true,
			wantOnline: false,
			numOfOnline: 0,
		},
		{
			name: "Test with both negative Timeout and Interval",
			args: args{
				CIDR:       "127.0.0.0/29",
				Count:      1,
				Timeout:    -2 * time.Second,
				Interval:   -500 * time.Millisecond,
				MaxWorkers: 2,
			},
			wantErr:    true,
			wantOnline: false,
			numOfOnline: 0,
		},
		{
			name: "Test with IPv6 ::1/128",
			args: args{
				CIDR:       "::1/128",
				Count:      4,
				Timeout:    1 * time.Second,
				Interval:   300 * time.Millisecond,
				MaxWorkers: 1,
			},
			wantErr:    false,
			wantOnline: true,
			numOfOnline: 1,
		},
		{
			name: "Test with IPv4 /20 should online all - high conccurency",
			args: args{
				CIDR:       "127.0.0.0/24",
				Count:      1,
				Timeout:    1 * time.Second,
				Interval:   300 * time.Millisecond,
				MaxWorkers: 256,
			},
			wantErr:    false,
			wantOnline: true,
			numOfOnline: 255,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp, err := subping.NewSubping(&subping.Options{
				Subnet:     tt.args.CIDR,
				Count:      tt.args.Count,
				Interval:   tt.args.Interval,
				Timeout:    tt.args.Timeout,
				MaxWorkers: tt.args.MaxWorkers,
				LogLevel:   "error",
			})
			if err != nil {
				if !tt.wantErr {
					t.Errorf("NewSubping() error = %v, wantErr %v", err, tt.wantErr)
				}

				return
			}

			if sp.TargetsIterator == nil && !tt.wantErr {
				t.Errorf("NewSubping() TargetsIterator = nil, want *network.SubnetHostsIterator, wantErr %v", tt.wantErr)
				return
			}

			if sp.Count != tt.args.Count {
				t.Errorf("NewSubping() Count got = %v, want %v, wantErr %v", sp.Count, tt.args.Count, tt.wantErr)
				return
			}

			if sp.Timeout != tt.args.Timeout {
				t.Errorf("NewSubping() Timeout got = %v, want %v, wantErr %v", sp.Timeout, tt.args.Timeout, tt.wantErr)
				return
			}

			if sp.Interval != tt.args.Interval {
				t.Errorf("NewSubping() Interval got = %v, want %v, wantErr %v", sp.Interval, tt.args.Interval, tt.wantErr)
				return
			}

			if sp.MaxWorkers != tt.args.MaxWorkers {
				t.Errorf("NewSubping() MaxWorkers got = %v, want %v, wantErr %v", sp.MaxWorkers, tt.args.MaxWorkers, tt.wantErr)
				return
			}

			sp.Run()
			_, onlineResultsLen := sp.GetOnlineHosts()

			wantTotalResults, err := network.CalculateTotalHostsFromCIDRString(tt.args.CIDR)
			if err != nil {
				t.Errorf("CalculateTotalHostsFromCIDRString() error => %v, wantErr %v", err, tt.wantErr)
				return
			}

			if wantTotalResults != sp.TotalResults {
				var hosts []string
				for k := range sp.Results {
					hosts = append(hosts, k)
				}

				sort.Strings(hosts)

				t.Errorf("Subping.Results length is invalid => got %v (%v), want %v, wantErr %v\nError: %v", sp.TotalResults, len(sp.Results), wantTotalResults, tt.wantErr, hosts)
				return
			}

			if tt.wantOnline && onlineResultsLen != tt.numOfOnline && onlineResultsLen != wantTotalResults {
				var hosts []string
				for k := range sp.Results {
					hosts = append(hosts, k)
				}

				t.Errorf("Subping.Results length online hosts is invalid => got %v, want %v, wantErr %v\n", onlineResultsLen, tt.numOfOnline, tt.wantErr)
				return
			}
		})
	}
}

func TestRunPing(t *testing.T) {
	type args struct {
		ipAddress string
		count     int
		interval  time.Duration
		timeout   time.Duration
	}

	type want struct {
		PacketsSent int
		PacketsRecv int
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test with valid local IP Address with Count 5",
			args: args{
				ipAddress: "localhost",
				count:     5,
				interval:  300 * time.Millisecond,
				timeout:   3 * time.Second,
			},
			want: want{
				PacketsSent: 5,
				PacketsRecv: 5,
			},
		},
		{
			name: "Test with invalid local IP Address with Count 5",
			args: args{
				ipAddress: "1",
				count:     5,
				interval:  300 * time.Millisecond,
				timeout:   3 * time.Second,
			},
			want: want{
				PacketsSent: 0,
				PacketsRecv: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := subping.RunPing(tt.args.ipAddress, tt.args.count, tt.args.interval, tt.args.timeout)

			if got.PacketsRecv != tt.want.PacketsRecv {
				t.Errorf("RunPing() PacketsRecv = %v, want %v", got.PacketsRecv, tt.want.PacketsRecv)
			}

			if got.PacketsSent != tt.want.PacketsSent {
				t.Errorf("RunPing() PacketsSent = %v, want %v", got.PacketsSent, tt.want.PacketsSent)
			}
		})
	}
}

// BenchmarkSubpingPerformance benchmarks the performance of Subping with various subnet sizes
func BenchmarkSubpingPerformance(b *testing.B) {
	tests := []struct {
		name       string
		cidr       string
		maxWorkers int
		count      int
		interval   time.Duration
		timeout    time.Duration
	}{
		{
			name:       "Small subnet /28",
			cidr:       "127.0.0.0/28",
			maxWorkers: 4,
			count:      1,
			interval:   100 * time.Millisecond,
			timeout:    500 * time.Millisecond,
		},
		{
			name:       "Medium subnet /24",
			cidr:       "127.0.0.0/24",
			maxWorkers: 32,
			count:      1,
			interval:   100 * time.Millisecond,
			timeout:    500 * time.Millisecond,
		},
		{
			name:       "Large subnet /20",
			cidr:       "127.0.0.0/20",
			maxWorkers: 64,
			count:      1,
			interval:   100 * time.Millisecond,
			timeout:    500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			s, err := subping.NewSubping(&subping.Options{
				Subnet:     tt.cidr,
				Count:      tt.count,
				Interval:   tt.interval,
				Timeout:    tt.timeout,
				MaxWorkers: tt.maxWorkers,
				LogLevel:   "error",
			})
			if err != nil {
				b.Fatalf("Failed to create Subping: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s.Results = nil // Reset results
				s.Run()
			}
		})
	}
}

// BenchmarkChannelBufferSize compares different channel buffer sizes
func BenchmarkChannelBufferSize(b *testing.B) {
	// This benchmark demonstrates the impact of channel buffer size optimization
	s, err := subping.NewSubping(&subping.Options{
		Subnet:     "127.0.0.0/24",
		Count:      1,
		Interval:   100 * time.Millisecond,
		Timeout:    500 * time.Millisecond,
		MaxWorkers: 128,
		LogLevel:   "error",
	})
	if err != nil {
		b.Fatalf("Failed to create Subping: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Results = nil // Reset results
		s.Run()
	}
}

// BenchmarkResultsMapAllocation benchmarks the results map allocation performance
func BenchmarkResultsMapAllocation(b *testing.B) {
	sizes := []struct {
		name string
		cidr string
	}{
		{"Small", "127.0.0.0/28"},
		{"Medium", "127.0.0.0/24"},
		{"Large", "127.0.0.0/20"},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			s, err := subping.NewSubping(&subping.Options{
				Subnet:     size.cidr,
				Count:      1,
				Interval:   100 * time.Millisecond,
				Timeout:    500 * time.Millisecond,
				MaxWorkers: 64,
				LogLevel:   "error",
			})
			if err != nil {
				b.Fatalf("Failed to create Subping: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s.Results = nil // Reset results
				s.Run()
			}
		})
	}
}
