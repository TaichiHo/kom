package utils

import (
	"fmt"
	"net"
)

func CidrTotalIPs(cidr string) (int, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, fmt.Errorf("invalid CIDR: %v", err)
	}

	// Get the mask size of the CIDR
	ones, bits := ipNet.Mask.Size()

	// Calculate total number of IPs
	totalIPs := 1 << (bits - ones)
	return totalIPs, nil
}
