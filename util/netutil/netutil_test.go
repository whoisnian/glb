package netutil_test

import (
	"net/netip"
	"testing"

	"github.com/whoisnian/glb/util/netutil"
)

func TestGetOutBoundIP(t *testing.T) {
	got, err := netutil.GetOutBoundIP()
	if err != nil {
		t.Fatalf("GetOutBoundIP() error: %v", err)
	}
	ip, ok := netip.AddrFromSlice(got)
	if !ok || !ip.IsValid() {
		t.Fatalf("GetOutBoundIP() = %v, invalid net.IP", got)
	}
	if ip.IsUnspecified() || ip.IsLoopback() || ip.IsMulticast() {
		t.Fatalf("GetOutBoundIP() = %v, invalid outbound IP", got)
	}
	if !ip.Is4() {
		t.Fatalf("GetOutBoundIP() = %v, invalid IPv4", got)
	}
}
