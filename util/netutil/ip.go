package netutil

import (
	"net"
	"strings"
)

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
	host, port, found := strings.CutLast(addr, ":")
	if !found {
		return addr, ""
	}
	if len(host) > 1 && host[0] == '[' && host[len(host)-1] == ']' {
		host = host[1 : len(host)-1]
	}
	return host, port
}
