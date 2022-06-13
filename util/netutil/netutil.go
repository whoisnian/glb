// Package netutil implements some network utility functions.
package netutil

import "net"

// GetOutBoundIP get preferred outbound ip of current process.
//
// https://stackoverflow.com/a/37382208
func GetOutBoundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
