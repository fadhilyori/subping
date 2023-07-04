package network

import (
	"errors"
	"net"
)

// FindIPsOutsideSubnet returns a list of IP addresses from the given slice
// that are outside the specified subnet.
func FindIPsOutsideSubnet(ipAddresses []net.IP, subnet *net.IPNet) []net.IP {
	var outsideSubnetIPs []net.IP

	for _, ipAddress := range ipAddresses {
		if !subnet.Contains(ipAddress) {
			outsideSubnetIPs = append(outsideSubnetIPs, ipAddress)
		}
	}

	return outsideSubnetIPs
}

// GenerateIPListFromCIDRString parses the given CIDR string and generates a list
// of IP addresses within the specified range.
// The CIDR string should be in the form "ip/mask", e.g., "192.168.0.0/24".
func GenerateIPListFromCIDRString(cidr string) ([]net.IP, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return []net.IP{}, errors.New("Failed to parse CIDR notation: %v\n")
	}

	return GenerateIPListFromCIDR(ipNet), nil
}

// GenerateIPListFromCIDR generates a list of IP addresses within the specified range
// based on the given CIDR notation.
func GenerateIPListFromCIDR(cidr *net.IPNet) []net.IP {
	var ips []net.IP

	firstIP := make(net.IP, len(cidr.IP))
	copy(firstIP, cidr.IP)

	for ip := firstIP; cidr.Contains(ip); inc(ip) {
		newIP := make(net.IP, len(ip))
		copy(newIP, ip)
		ips = append(ips, newIP)
	}

	return append([]net.IP{}, ips...)
}

// inc increments the given IP address by one.
// It handles both IPv4 and IPv6 addresses.
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
