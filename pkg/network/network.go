package network

import (
	"errors"
	"net"
)

func FindIPsOutsideSubnet(ipAddresses []net.IP, subnet *net.IPNet) []net.IP {
	var outsideSubnetIPs []net.IP

	for _, ipAddress := range ipAddresses {
		if !subnet.Contains(ipAddress) {
			outsideSubnetIPs = append(outsideSubnetIPs, ipAddress)
		}
	}

	return outsideSubnetIPs
}

func GenerateIPListFromCIDRString(cidr string) ([]net.IP, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return []net.IP{}, errors.New("Failed to parse CIDR notation: %v\n")
	}

	return GenerateIPListFromCIDR(ip, ipNet), nil
}

func GenerateIPListFromCIDR(firstIp net.IP, cidr *net.IPNet) []net.IP {
	var ips []net.IP

	for ip := firstIp; cidr.Contains(ip); inc(ip) {
		newIP := make(net.IP, len(ip))
		copy(newIP, ip)
		ips = append(ips, newIP)
	}

	return append([]net.IP{}, ips...)
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
