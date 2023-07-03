package subping_test

import (
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/fadhilyori/subping"
	"github.com/go-ping/ping"
)

var (
	targetsLocalWithSubnet29 = []net.IP{
		net.ParseIP("127.0.0.1"),
		net.ParseIP("127.0.0.2"),
		net.ParseIP("127.0.0.3"),
		net.ParseIP("127.0.0.4"),
		net.ParseIP("127.0.0.5"),
		net.ParseIP("127.0.0.6"),
	}
)

func TestNewSubping(t *testing.T) {
	type args struct {
		opts *subping.Options
	}
	tests := []struct {
		name    string
		args    args
		want    subping.Subping
		wantErr bool
	}{
		{
			name: "Test with valid options",
			args: args{
				opts: &subping.Options{
					Targets:  targetsLocalWithSubnet29,
					Count:    3,
					Timeout:  1 * time.Second,
					Interval: 300 * time.Millisecond,
					NumJobs:  2,
				},
			},
			want: subping.Subping{
				Targets:  targetsLocalWithSubnet29,
				Count:    3,
				Timeout:  1 * time.Second,
				Interval: 300 * time.Millisecond,
				NumJobs:  2,
			},
			wantErr: false,
		},
		{
			name: "Test with invalid Count",
			args: args{
				opts: &subping.Options{
					Targets:  targetsLocalWithSubnet29,
					Count:    -1,
					Timeout:  1 * time.Second,
					Interval: 300 * time.Millisecond,
					NumJobs:  2,
				},
			},
			want: subping.Subping{
				Targets:  nil,
				Count:    0,
				Timeout:  0,
				Interval: 0,
				NumJobs:  0,
			},
			wantErr: true,
		},
		{
			name: "Test with invalid NumJobs",
			args: args{
				opts: &subping.Options{
					Targets:  targetsLocalWithSubnet29,
					Count:    1,
					Interval: 300 * time.Millisecond,
					NumJobs:  -2,
				},
			},
			want: subping.Subping{
				Targets:  nil,
				Count:    0,
				Interval: 0,
				NumJobs:  0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := subping.NewSubping(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSubping() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got.Targets, tt.want.Targets) {
				t.Errorf("NewSubping() Targets got = %v, want %v", got, tt.want)
			}

			if got.Count != tt.want.Count {
				t.Errorf("NewSubping() Count got = %v, want %v", got, tt.want)
			}

			if got.Timeout != tt.want.Timeout {
				t.Errorf("NewSubping() Timeout got = %v, want %v", got, tt.want)
			}

			if got.Interval != tt.want.Interval {
				t.Errorf("NewSubping() Interval got = %v, want %v", got, tt.want)
			}

			if got.NumJobs != tt.want.NumJobs {
				t.Errorf("NewSubping() NumJobs got = %v, want %v", got, tt.want)
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
				PacketsSent: 5,
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

func TestSubping_GetOnlineHosts(t *testing.T) {
	type fields struct {
		targets  []net.IP
		count    int
		interval time.Duration
		numJobs  int
		results  map[string]*ping.Statistics
	}

	var targetsLocalWithSubnet27 []net.IP

	for i := 1; i < 30; i++ {
		targetsLocalWithSubnet27 = append(targetsLocalWithSubnet27, net.ParseIP(fmt.Sprintf("127.0.0.%d", i)))
	}

	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Ping to 127.0.0.1/27 with number of jobs 2",
			fields: fields{
				targets:  targetsLocalWithSubnet27,
				count:    3,
				interval: 300 * time.Millisecond,
				numJobs:  2,
			},
		},
		{
			name: "Ping to 127.0.0.1/27 with number of jobs 20",
			fields: fields{
				targets:  targetsLocalWithSubnet27,
				count:    3,
				interval: 300 * time.Millisecond,
				numJobs:  20,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &subping.Subping{
				Targets:  tt.fields.targets,
				Count:    tt.fields.count,
				Interval: tt.fields.interval,
				NumJobs:  tt.fields.numJobs,
			}

			s.Run()

			got := s.GetOnlineHosts()

			for i, h := range got {
				if h.PacketsRecv == 0 {
					t.Errorf("GetOnlineHosts() = %v should be offline", i)
				}
			}
		})
	}
}

func TestSubping_Run(t *testing.T) {
	var targetsLocalWithSubnet27 []net.IP

	for i := 1; i < 30; i++ {
		targetsLocalWithSubnet27 = append(targetsLocalWithSubnet27, net.ParseIP(fmt.Sprintf("127.0.0.%d", i)))
	}

	type fields struct {
		targets  []net.IP
		count    int
		interval time.Duration
		timeout  time.Duration
		numJobs  int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Ping to 127.0.0.1/27 with number of jobs 2",
			fields: fields{
				targets:  targetsLocalWithSubnet27,
				count:    3,
				interval: 300 * time.Millisecond,
				timeout:  1 * time.Second,
				numJobs:  2,
			},
		},
		{
			name: "Ping to 127.0.0.1/27 with number of jobs 50",
			fields: fields{
				targets:  targetsLocalWithSubnet27,
				count:    3,
				interval: 300 * time.Millisecond,
				timeout:  1 * time.Second,
				numJobs:  50,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := subping.NewSubping(&subping.Options{
				Targets:  tt.fields.targets,
				Count:    tt.fields.count,
				Interval: tt.fields.interval,
				Timeout:  tt.fields.timeout,
				NumJobs:  tt.fields.numJobs,
			})
			s.Run()

			lenOfResults := len(s.Results)
			lenOfTargets := len(s.Targets)

			if lenOfResults != lenOfTargets {
				t.Errorf("Run() = %v, want %v", lenOfResults, lenOfTargets)
			}
		})
	}
}
