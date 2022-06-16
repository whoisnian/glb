package netutil

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"
)

var TEST_SEED int64

func init() {
	TEST_SEED = time.Now().Unix()
	fmt.Printf("Running with rand seed %v\n", TEST_SEED)
}

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
		if out := FirstIP(cidr); !out.Equal(tt.from) {
			t.Errorf("CIDR(%v).FirstIP = %v, want %v", tt.in, out, tt.from)
		}
	}
}

func TestLastIP(t *testing.T) {
	for _, tt := range cidrRangeTests {
		_, cidr, _ := net.ParseCIDR(tt.in)
		if out := LastIP(cidr); !out.Equal(tt.to) {
			t.Errorf("CIDR(%v).LastIP = %v, want %v", tt.in, out, tt.to)
		}
	}
}

type SimpleList struct {
	list []net.IPNet
}

func NewSimpleList() *SimpleList {
	return &SimpleList{}
}

func (s *SimpleList) Add(cidr net.IPNet) {
	s.list = append(s.list, cidr)
}

func (s *SimpleList) Contains(ip net.IP) bool {
	for _, cidr := range s.list {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

var LIST_SIZE = 36

func TestChecklist4(t *testing.T) {
	r := rand.New(rand.NewSource(TEST_SEED))
	simplelist := NewSimpleList()
	checklist := NewChecklist4()
	for i := 0; i < LIST_SIZE; i++ {
		buf := make([]byte, net.IPv4len)
		r.Read(buf)
		cidr := net.IPNet{IP: buf, Mask: net.CIDRMask(r.Intn(25)+8, 32)}
		simplelist.Add(cidr)
		checklist.Add(cidr)
	}

	buf := make([]byte, net.IPv4len)
	for i := 0; i < 1e6; i++ {
		r.Read(buf)
		if simplelist.Contains(buf) != checklist.Contains(buf) {
			t.Errorf("Checklist4.Contains(%v) = %v, want %v", buf, checklist.Contains(buf), simplelist.Contains(buf))
		}
	}
}

func BenchmarkSimpleList(b *testing.B) {
	r := rand.New(rand.NewSource(TEST_SEED))
	simplelist := NewSimpleList()
	for i := 0; i < LIST_SIZE; i++ {
		buf := make([]byte, net.IPv4len)
		r.Read(buf)
		cidr := net.IPNet{IP: buf, Mask: net.CIDRMask(r.Intn(25)+8, 32)}
		simplelist.Add(cidr)
	}

	buf := make([]byte, net.IPv4len)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Read(buf)
		simplelist.Contains(buf)
	}
}

func BenchmarkChecklist4(b *testing.B) {
	r := rand.New(rand.NewSource(TEST_SEED))
	checklist := NewChecklist4()
	for i := 0; i < LIST_SIZE; i++ {
		buf := make([]byte, net.IPv4len)
		r.Read(buf)
		cidr := net.IPNet{IP: buf, Mask: net.CIDRMask(r.Intn(25)+8, 32)}
		checklist.Add(cidr)
	}

	buf := make([]byte, net.IPv4len)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Read(buf)
		checklist.Contains(buf)
	}
}
