// Package network provides functionality for working with IP networks and subnet hosts.
//
// The package includes functions for iterating over hosts within a subnet, calculating the total number of hosts
// in a subnet, parsing CIDR notation, and obtaining the first and last IP addresses from an IP network.
//
// Examples:
//
//	ipNet := &net.IPNet{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(24, 32)}
//	it := network.NewSubnetHostsIterator(ipNet)
//
//	for ip := it.Next(); ip != nil; ip = it.Next() {
//		// Process the host IP
//		fmt.Println(ip.String())
//	}
//
//	cidr := "192.168.0.0/24"
//	totalHosts, err := network.CalculateTotalHostsFromCIDRString(cidr)
//	if err != nil {
//		fmt.Println("Error:", err)
//		return
//	}
//	fmt.Println("Total hosts:", totalHosts)
//
//	firstIP := network.GetFirstIPAddressFromIPNet(ipNet)
//	lastIP := network.GetLastIPAddressFromIPNet(ipNet)
//	fmt.Println("First IP:", firstIP)
//	fmt.Println("Last IP:", lastIP)
package network

import (
	"errors"
	"math"
	"net"
	"sync"
)

// SubnetHostsIterator represents an iterator over the hosts within a subnet.
type SubnetHostsIterator struct {
	// IPNet represents the subnet to iterate over.
	IPNet *net.IPNet

	// CurrentIP represents the current host IP.
	CurrentIP *net.IP

	// FirstIP represents the first host IP in the subnet.
	FirstIP net.IP

	// LastIP represents the last host IP in the subnet.
	LastIP net.IP

	// TotalHosts represents the total number of hosts in the subnet.
	TotalHosts int

	// mu is a mutex used for thread-safety.
	mu sync.Mutex
}

// NewSubnetHostsIteratorFromCIDRString creates a new SubnetHostsIterator for the given CIDR string.
// It parses the CIDR string, creates an IP network, and initializes the iterator with the necessary values.
func NewSubnetHostsIteratorFromCIDRString(cidr string) (*SubnetHostsIterator, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, errors.New("failed to parse CIDR notation")
	}

	return NewSubnetHostsIterator(ipNet), nil
}

// NewSubnetHostsIterator creates a new SubnetHostsIterator for the given IP network.
// It initializes the iterator with the first and last host IPs, the IP network, the current IP,
// and the total number of hosts in the subnet.
func NewSubnetHostsIterator(ipNet *net.IPNet) *SubnetHostsIterator {
	return &SubnetHostsIterator{
		FirstIP:    GetFirstIPAddressFromIPNet(ipNet),
		LastIP:     GetLastIPAddressFromIPNet(ipNet),
		IPNet:      ipNet,
		CurrentIP:  nil,
		TotalHosts: CalculateTotalHosts(ipNet),
	}
}

// Next returns the next host IP in the subnet. It locks the iterator for thread-safety.
// If it's the first call to Next, it returns the first host IP in the subnet.
// If there are no more hosts in the subnet or if the current IP is outside the subnet,
// it returns nil.
func (it *SubnetHostsIterator) Next() *net.IP {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.CurrentIP == nil {
		currentIP := make(net.IP, len(it.FirstIP))
		copy(currentIP, it.FirstIP)
		it.CurrentIP = &currentIP
		return it.CurrentIP
	}

	currentIP := *it.CurrentIP

	for i := len(currentIP) - 1; i >= 0; i-- {
		currentIP[i]++
		if currentIP[i] > 0 {
			break
		}
	}

	if !it.IPNet.Contains(currentIP) {
		return nil
	}

	return &currentIP
}

// GetFirstIPAddressFromIPNet returns the first host IP address within the given IP network.
func GetFirstIPAddressFromIPNet(ipNet *net.IPNet) net.IP {
	firstIP := make(net.IP, len(ipNet.IP))
	copy(firstIP, ipNet.IP)

	return firstIP
}

// GetLastIPAddressFromIPNet returns the last host IP address within the given IP network.
func GetLastIPAddressFromIPNet(ipNet *net.IPNet) net.IP {
	lastIP := make(net.IP, len(ipNet.IP))
	copy(lastIP, ipNet.IP)
	for i := range lastIP {
		lastIP[i] |= ^ipNet.Mask[i]
	}

	return lastIP
}

// CalculateTotalHostsFromCIDRString calculates the total number of hosts based on the provided CIDR string.
func CalculateTotalHostsFromCIDRString(cidr string) (int, error) {
	_, parsedCIDR, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, err
	}

	return CalculateTotalHosts(parsedCIDR), nil
}

// CalculateTotalHosts calculates the total number of hosts based on the provided IP network.
func CalculateTotalHosts(ipNet *net.IPNet) int {
	// Calculate the number of host bits
	prefixLength, totalBits := ipNet.Mask.Size()
	hostBits := totalBits - prefixLength

	// Calculate the total hosts based on the number of host bits
	totalHosts := int(math.Pow(2, float64(hostBits)))

	return totalHosts
}
