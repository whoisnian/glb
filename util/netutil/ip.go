package netutil

import "net"

// FirstIP returns the start IP of specified CIDR.
func FirstIP(cidr *net.IPNet) net.IP {
	return cidr.IP.Mask(cidr.Mask)
}

// LastIP returns the end IP of specified CIDR.
func LastIP(cidr *net.IPNet) net.IP {
	ip := cidr.IP.Mask(cidr.Mask)
	for i := len(cidr.Mask) - 1; i >= 0; i-- {
		ip[i] = ip[i] | ^cidr.Mask[i]
	}
	return ip
}
