package network_test

import (
	"math"
	"net"
	"testing"

	"github.com/fadhilyori/subping/pkg/network"
)

func TestHostsIterator(t *testing.T) {
	tests := []struct {
		name string
		cidr string
		want int
	}{
		{
			name: "IPv4 Subnet 28",
			cidr: "127.0.0.0/28",
			want: int(math.Pow(2, 32-28)),
		},
		{
			name: "IPv4 Subnet 24",
			cidr: "127.0.0.0/24",
			want: int(math.Pow(2, 32-24)),
		},
		{
			name: "IPv4 Subnet 8",
			cidr: "127.0.0.0/8",
			want: int(math.Pow(2, 32-8)),
		},
		{
			name: "IPv6 Subnet 125",
			cidr: "2001:db8:1::/125",
			want: int(math.Pow(2, 128-125)),
		},
		{
			name: "IPv6 Subnet 120",
			cidr: "2001:db8:1::/120",
			want: int(math.Pow(2, 128-120)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			iterator, err := network.NewSubnetHostsIteratorFromCIDRString(tt.cidr)
			if err != nil {
				t.Errorf("NewSubnetHostsIteratorFromString() error => %v", err)
			}
			_, ipNet, _ := net.ParseCIDR(tt.cidr)

			for ip := iterator.Next(); ip != nil; ip = iterator.Next() {
				if !ipNet.Contains(*ip) {
					t.Errorf("Next() host should not in the subnet %s, got %s", ipNet.String(), ip.String())
				}
				count++
			}

			if count != tt.want || count != iterator.TotalHosts {
				t.Errorf("SubnetHostsIterator{} number of hosts is not %d, got %d (%d)", tt.want, count, iterator.TotalHosts)
			}
		})
	}
}

func BenchmarkHostsIterator(b *testing.B) {
	tests := []struct {
		name string
		cidr string
		want int
	}{
		{
			name: "IPv6 Subnet 100",
			cidr: "2001:db8:1::/100",
			want: int(math.Pow(2, 128-100)),
		},
	}
	for _, bb := range tests {
		b.Run(bb.name, func(b *testing.B) {
			count := 0
			iterator, err := network.NewSubnetHostsIteratorFromCIDRString(bb.cidr)
			if err != nil {
				b.Errorf("NewSubnetHostsIteratorFromString() error => %v", err)
			}

			_, ipNet, _ := net.ParseCIDR(bb.cidr)

			for ip := iterator.Next(); ip != nil; ip = iterator.Next() {
				if !ipNet.Contains(*ip) {
					b.Errorf("Next() host should not in the subnet %s, got %s", ipNet.String(), ip.String())
				}
				count++
			}

			if count != bb.want || count != iterator.TotalHosts {
				b.Errorf("SubnetHostsIterator{} number of hosts is not %d, got %d (%d)", bb.want, count, iterator.TotalHosts)
			}
		})
	}
}

func TestCalculateTotalHostsFromCIDRString(t *testing.T) {
	type args struct {
		cidr string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "IPv4 /24",
			args: args{
				cidr: "127.0.0.0/24",
			},
			want:    256,
			wantErr: false,
		},
		{
			name: "IPv6 /64",
			args: args{
				cidr: "::1/120",
			},
			want:    256,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := network.CalculateTotalHostsFromCIDRString(tt.args.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateTotalHostsFromCIDRString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CalculateTotalHostsFromCIDRString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
