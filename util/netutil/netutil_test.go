package netutil

import (
	"net/netip"
	"testing"
)

func TestGetOutBoundIP(t *testing.T) {
	got, err := GetOutBoundIP()
	if err != nil {
		t.Errorf("GetOutBoundIP() error: %v", err)
	}
	ip, ok := netip.AddrFromSlice(got)
	if !ok || !ip.IsValid() {
		t.Errorf("GetOutBoundIP() = %v, invalid net.IP", got)
	}
	if ip.IsUnspecified() || ip.IsLoopback() || ip.IsMulticast() {
		t.Errorf("GetOutBoundIP() = %v, invalid outbound IP", got)
	}
	if !ip.Is4() {
		t.Errorf("GetOutBoundIP() = %v, invalid IPv4", got)
	}
}
