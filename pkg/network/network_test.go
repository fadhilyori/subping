package network

import (
	"net"
	"testing"
)

func TestFindIPsOutsideSubnet(t *testing.T) {
	subnet := &net.IPNet{
		IP:   net.ParseIP("192.168.0.0"),
		Mask: net.CIDRMask(24, 32),
	}

	ipAddresses := []net.IP{
		net.ParseIP("192.168.0.1"),
		net.ParseIP("192.168.0.2"),
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}

	expectedOutsideSubnetIPs := []net.IP{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
	}

	outsideSubnetIPs := FindIPsOutsideSubnet(ipAddresses, subnet)

	if len(outsideSubnetIPs) != len(expectedOutsideSubnetIPs) {
		t.Errorf("Unexpected number of IP addresses outside subnet. Expected: %d, Got: %d",
			len(expectedOutsideSubnetIPs), len(outsideSubnetIPs))
	}

	for i, ip := range outsideSubnetIPs {
		if !ip.Equal(expectedOutsideSubnetIPs[i]) {
			t.Errorf("IP address outside subnet does not match the expected value. Expected: %s, Got: %s",
				expectedOutsideSubnetIPs[i].String(), ip.String())
		}
	}
}

func TestGenerateIPListFromCIDR30(t *testing.T) {
	ip := net.ParseIP("192.168.0.1")
	_, cidr, _ := net.ParseCIDR("192.168.0.0/30")

	result := GenerateIPListFromCIDR(ip, cidr)

	checkResult := FindIPsOutsideSubnet(result, cidr)
	if len(checkResult) != 0 {
		t.Errorf("Generated IP list contains invalid IPs.\nInvalid IPs: %v\nGot: %v", checkResult, result)
	}
}
