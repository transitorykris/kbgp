package kbgp

import (
	"encoding/binary"
	"fmt"
	"net"
)

// FindBGPIdentifier tries to find the best possible identifier from all
// interfaces configured on the host
func FindBGPIdentifier() (uint32, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return 0, err
	}
	// Note: this selection process is arbitrary
	for _, v := range ifs {
		addrs, err := v.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}
			// Make sure we have an IPv4 address
			if ip.To4() == nil {
				continue
			}
			// If it's routable, we have a winner!
			if ip.IsGlobalUnicast() {
				return ipToUint32(ip), nil
			}
		}
	}
	return 0, fmt.Errorf("No valid BGP identifier found")
}

func ipToUint32(ip net.IP) uint32 {
	// ip could be 4 or 16 bytes, let's be sure it's 4
	ip4 := ip.To4()
	u := binary.BigEndian.Uint32(ip4)
	return u
}

// Uint32ToIP converts a uint32 to a net.IP
func Uint32ToIP(i uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, i)
	return ip
}
