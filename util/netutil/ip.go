package netutil

import "net"

func FirstIP(cidr *net.IPNet) net.IP {
	return cidr.IP.Mask(cidr.Mask)
}

func LastIP(cidr *net.IPNet) net.IP {
	ip := cidr.IP.Mask(cidr.Mask)
	for i := len(cidr.Mask) - 1; i >= 0; i-- {
		ip[i] = ip[i] | ^cidr.Mask[i]
	}
	return ip
}
