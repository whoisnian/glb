package netutil_test

import (
	"net"
	"testing"

	"github.com/whoisnian/glb/util/netutil"
)

var cidrRangeTests = []struct {
	in   string
	from net.IP
	to   net.IP
}{
	{"192.168.0.1/12", net.ParseIP("192.160.0.0"), net.ParseIP("192.175.255.255")},
	{"192.168.0.1/16", net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},
	{"192.168.0.1/20", net.ParseIP("192.168.0.0"), net.ParseIP("192.168.15.255")},
	{"192.168.5.1/22", net.ParseIP("192.168.4.0"), net.ParseIP("192.168.7.255")},
	{"192.168.5.1/24", net.ParseIP("192.168.5.0"), net.ParseIP("192.168.5.255")},
	{"192.168.5.1/32", net.ParseIP("192.168.5.1"), net.ParseIP("192.168.5.1")},
	{"abcd:2300::/24", net.ParseIP("abcd:2300::"), net.ParseIP("abcd:23ff:ffff:ffff:ffff:ffff:ffff:ffff")},
	{"abcd:2345::/24", net.ParseIP("abcd:2300::"), net.ParseIP("abcd:23ff:ffff:ffff:ffff:ffff:ffff:ffff")},
	{"abcd:2344::/31", net.ParseIP("abcd:2344::"), net.ParseIP("abcd:2345:ffff:ffff:ffff:ffff:ffff:ffff")},
	{"abcd:2345::/32", net.ParseIP("abcd:2345::"), net.ParseIP("abcd:2345:ffff:ffff:ffff:ffff:ffff:ffff")},
	{"abcd:2345::/33", net.ParseIP("abcd:2345::"), net.ParseIP("abcd:2345:7fff:ffff:ffff:ffff:ffff:ffff")},
	{"abcd:2345::/63", net.ParseIP("abcd:2345::"), net.ParseIP("abcd:2345:0:1:ffff:ffff:ffff:ffff")},
	{"abcd:2345::/64", net.ParseIP("abcd:2345::"), net.ParseIP("abcd:2345:0:0:ffff:ffff:ffff:ffff")},
	{"abcd:2345::/65", net.ParseIP("abcd:2345::"), net.ParseIP("abcd:2345:0:0:7fff:ffff:ffff:ffff")},
	{"abcd:2345::/127", net.ParseIP("abcd:2345::"), net.ParseIP("abcd:2345:0:0:0:0:0:1")},
	{"::1/128", net.ParseIP("::1"), net.ParseIP("::1")},
}

func TestFirstIP(t *testing.T) {
	for _, tt := range cidrRangeTests {
		_, cidr, _ := net.ParseCIDR(tt.in)
		if out := netutil.FirstIP(cidr); !out.Equal(tt.from) {
			t.Errorf("CIDR(%v).FirstIP = %v, want %v", tt.in, out, tt.from)
		}
	}
}

func TestLastIP(t *testing.T) {
	for _, tt := range cidrRangeTests {
		_, cidr, _ := net.ParseCIDR(tt.in)
		if out := netutil.LastIP(cidr); !out.Equal(tt.to) {
			t.Errorf("CIDR(%v).LastIP = %v, want %v", tt.in, out, tt.to)
		}
	}
}

func TestSplitHostPort(t *testing.T) {
	var tests = []struct {
		addr string
		host string
		port string
	}{
		{"127.0.0.1:80", "127.0.0.1", "80"},
		{"[::1]:80", "::1", "80"},
		{"127.0.0.1:", "127.0.0.1", ""},
		{"[::1]:", "::1", ""},
		{"127.0.0.1", "127.0.0.1", ""},
		{":80", "", "80"},
		{":", "", ""},
		{"", "", ""},
	}
	for _, test := range tests {
		host, port := netutil.SplitHostPort(test.addr)
		if host != test.host || port != test.port {
			t.Fatalf("splitHostPort(%s) = (%s,%s), want (%s,%s)", test.addr, host, port, test.host, test.port)
		}
	}
}
