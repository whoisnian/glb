package netutil_test

import (
	"net/netip"
	"testing"

	"github.com/whoisnian/glb/util/netutil"
)

func TestGetOutBoundIP(t *testing.T) {
	got, err := netutil.GetOutBoundIP()
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
