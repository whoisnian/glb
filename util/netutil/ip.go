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

// SplitHostPort splits "host:port" or "[host]:port" into host and port without strict validation.
func SplitHostPort(addr string) (host, port string) {
	i := len(addr) - 1
	for ; i >= 0; i-- {
		if addr[i] == ':' {
			if addr[0] == '[' && addr[i-1] == ']' {
				return addr[1 : i-1], addr[i+1:]
			} else {
				return addr[0:i], addr[i+1:]
			}
		}
	}
	return addr, ""
}
