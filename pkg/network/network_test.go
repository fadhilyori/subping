package network_test

import (
	"net"
	"reflect"
	"testing"

	"github.com/fadhilyori/subping/pkg/network"
)

func TestFindIPsOutsideSubnet(t *testing.T) {
	type args struct {
		ipAddresses []net.IP
		subnet      *net.IPNet
	}
	tests := []struct {
		name string
		args args
		want []net.IP
	}{
		{
			name: "Should be no invalid IP Address in subnet 24",
			args: args{
				ipAddresses: []net.IP{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
					net.ParseIP("192.168.1.3"),
				},
				subnet: &net.IPNet{
					IP:   net.ParseIP("192.168.1.0"),
					Mask: net.CIDRMask(24, 32),
				},
			},
			want: []net.IP{},
		},
		{
			name: "Should be 1 invalid IP Address in subnet 24",
			args: args{
				ipAddresses: []net.IP{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
					net.ParseIP("192.168.1.3"),
					net.ParseIP("192.168.2.5"),
				},
				subnet: &net.IPNet{
					IP:   net.ParseIP("192.168.1.0"),
					Mask: net.CIDRMask(24, 32),
				},
			},
			want: []net.IP{
				net.ParseIP("192.168.2.5"),
			},
		},
		{
			name: "Should be 3 invalid IP Address in subnet 28",
			args: args{
				ipAddresses: []net.IP{
					net.ParseIP("192.168.1.1"),
					net.ParseIP("192.168.1.2"),
					net.ParseIP("192.168.1.3"),
					net.ParseIP("192.168.1.16"),
					net.ParseIP("192.168.1.65"),
					net.ParseIP("192.168.2.4"),
				},
				subnet: &net.IPNet{
					IP:   net.ParseIP("192.168.1.0"),
					Mask: net.CIDRMask(28, 32),
				},
			},
			want: []net.IP{
				net.ParseIP("192.168.1.16"),
				net.ParseIP("192.168.1.65"),
				net.ParseIP("192.168.2.4"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := network.FindIPsOutsideSubnet(tt.args.ipAddresses, tt.args.subnet); !reflect.DeepEqual(got, tt.want) {
				if len(got) == 0 && len(tt.want) == 0 {
					return
				}

				t.Errorf("FindIPsOutsideSubnet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateIPListFromCIDR(t *testing.T) {
	type args struct {
		cidr *net.IPNet
	}

	_, subnet16, _ := net.ParseCIDR("192.168.1.0/16")
	_, subnet24, _ := net.ParseCIDR("192.168.1.0/24")
	_, subnet30, _ := net.ParseCIDR("192.168.1.0/30")

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Generate IP list for a /16 subnet should have 65536 entries.",
			args: args{
				cidr: subnet16,
			},
		},
		{
			name: "Generate IP list for a /24 subnet should have 256 entries.",
			args: args{
				cidr: subnet24,
			},
		},
		{
			name: "Generate IP list for a /30 subnet should have 4 entries.",
			args: args{
				cidr: subnet30,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := network.GenerateIPListFromCIDR(tt.args.cidr)

			var wrongIPS []string

			for _, ip := range got {
				if !tt.args.cidr.Contains(ip) {
					wrongIPS = append(wrongIPS, ip.String())
				}
			}

			if len(wrongIPS) > 0 {
				t.Errorf("GenerateIPListFromCIDR() invalid IP = %v", wrongIPS)
			}
		})
	}
}

func TestGenerateIPListFromCIDRString(t *testing.T) {
	type args struct {
		cidr string
	}

	tests := []struct {
		name    string
		args    args
		want    []net.IP
		wantErr bool
	}{
		{
			name: "Generate IP list for a /16 subnet should have 65536 entries.",
			args: args{
				cidr: "192.168.1.0/16",
			},
			wantErr: false,
		},
		{
			name: "Generate IP list for a /24 subnet should have 256 entries.",
			args: args{
				cidr: "192.168.1.0/24",
			},
			wantErr: false,
		},
		{
			name: "Generate IP list for a /30 subnet should have 4 entries.",
			args: args{
				cidr: "192.168.1.0/30",
			},
			wantErr: false,
		},
		{
			name: "Generate IP list for a invalid subnet should error.",
			args: args{
				cidr: "192.168.1.0",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := network.GenerateIPListFromCIDRString(tt.args.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateIPListFromCIDRString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			_, cidr, _ := net.ParseCIDR(tt.args.cidr)

			var wrongIPS []string

			for _, ip := range got {
				if !cidr.Contains(ip) {
					wrongIPS = append(wrongIPS, ip.String())
				}
			}

			if len(wrongIPS) > 0 {
				t.Errorf("GenerateIPListFromCIDR() invalid IP = %v", wrongIPS)
			}
		})
	}
}
