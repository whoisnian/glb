package netutil

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

var TEST_SEED int64

func init() {
	TEST_SEED = time.Now().Unix()
	fmt.Printf("Running tests with rand seed %v\n", TEST_SEED)
}

type SimpleIPNetList struct {
	matchAll uint32
	ipList   []net.IPNet
}

func (s *SimpleIPNetList) Add(cidr *net.IPNet) error {
	if ones, _ := cidr.Mask.Size(); ones == 0 {
		atomic.StoreUint32(&s.matchAll, 1)
		return nil
	}
	s.ipList = append(s.ipList, net.IPNet{IP: cidr.IP.Mask(cidr.Mask), Mask: append([]byte(nil), cidr.Mask...)})
	return nil
}
func (s *SimpleIPNetList) Remove(cidr *net.IPNet) error {
	if ones, _ := cidr.Mask.Size(); ones == 0 {
		atomic.StoreUint32(&s.matchAll, 0)
		return nil
	}
	for i := range s.ipList {
		if len(s.ipList[i].Mask) > 0 && bytes.Equal(cidr.Mask, s.ipList[i].Mask) && cidr.IP.Mask(cidr.Mask).Equal(s.ipList[i].IP) {
			s.ipList[i] = net.IPNet{IP: net.IP{0, 0, 0, 0}, Mask: []byte{}} // reset to invalid CIDR
		}
	}
	return nil
}
func (s *SimpleIPNetList) Contains(ip net.IP) bool {
	if atomic.LoadUint32(&s.matchAll) == 1 {
		return true
	}
	for _, cidr := range s.ipList {
		if len(cidr.Mask) > 0 && cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func testIPv4Filter(t *testing.T, size int) {
	r := rand.New(rand.NewSource(TEST_SEED))
	simple := &SimpleIPNetList{}
	filter := NewIPv4Filter()
	delList := []*net.IPNet{}
	for i := 0; i < size; i++ {
		buf := make([]byte, net.IPv4len)
		r.Read(buf)
		cidr := &net.IPNet{IP: buf, Mask: net.CIDRMask(r.Intn(23)+10, 32)}
		simple.Add(cidr)
		filter.Add(cidr)
		if i%3 == 0 {
			delList = append(delList, cidr)
		}
	}
	for _, cidr := range delList {
		simple.Remove(cidr)
		filter.Remove(cidr)
	}

	buf := make([]byte, net.IPv4len)
	for i := 0; i < 1e5; i++ {
		r.Read(buf)
		if simple.Contains(buf) != filter.Contains(buf) {
			t.Log(simple.ipList)
			t.Log(filter.ipMaps)
			t.Fatalf("IPv4Filter.Contains(%v) = %v, want %v", buf, filter.Contains(buf), simple.Contains(buf))
		}
	}
}

func TestIPv4FilterModeList(t *testing.T) {
	testIPv4Filter(t, listSize)
}

func TestIPv4FilterModeMaps(t *testing.T) {
	testIPv4Filter(t, listSize*2)
}

func TestIPv4FilterMatchAll(t *testing.T) {
	cidrAll := &net.IPNet{IP: net.IP{0, 0, 0, 0}, Mask: net.CIDRMask(0, 32)}
	testIP := net.IPv4(192, 168, 0, 1)

	filter := NewIPv4Filter()
	if filter.Contains(testIP) {
		t.Fatalf("IPv4Filter.Contains(%v) = %v, want %v", testIP, true, false)
	}

	if err := filter.Add(cidrAll); err != nil {
		t.Log(cidrAll.Mask.Size())
		t.Log(len(cidrAll.IP), net.IPv4len)
		t.Fatal(err)
	}
	if !filter.Contains(testIP) {
		t.Log(cidrAll.Mask.Size())
		t.Fatalf("IPv4Filter.Contains(%v) = %v, want %v", testIP, false, true)
	}

	filter.Remove(cidrAll)
	if filter.Contains(testIP) {
		t.Fatalf("IPv4Filter.Contains(%v) = %v, want %v", testIP, true, false)
	}
}

type Filter interface {
	Add(*net.IPNet) error
	Remove(*net.IPNet) error
	Contains(net.IP) bool
}

func benchmarkFilter(b *testing.B, f Filter, size int) {
	r := rand.New(rand.NewSource(TEST_SEED))
	for i := 0; i < size; i++ {
		buf := make([]byte, net.IPv4len)
		r.Read(buf)
		cidr := &net.IPNet{IP: buf, Mask: net.CIDRMask(r.Intn(25)+8, 32)}
		f.Add(cidr)
	}

	buf := make([]byte, net.IPv4len)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Read(buf)
		f.Contains(buf)
	}
}

func BenchmarkSimpleIPNetList32(b *testing.B) {
	simple := &SimpleIPNetList{}
	benchmarkFilter(b, simple, 32)
}

func BenchmarkIPv4FilterModeList32(b *testing.B) {
	filter := NewIPv4Filter()
	filter.mode = modeList
	benchmarkFilter(b, filter, 32)
}

func BenchmarkIPv4FilterModeMaps32(b *testing.B) {
	filter := NewIPv4Filter()
	filter.mode = modeMaps
	for i := 0; i < len(filter.ipMaps); i++ {
		filter.ipMaps[i] = make(map[uint32]bool)
	}
	benchmarkFilter(b, filter, 32)
}

func BenchmarkSimpleIPNetList128(b *testing.B) {
	simple := &SimpleIPNetList{}
	benchmarkFilter(b, simple, 128)
}

func BenchmarkIPv4FilterModeList128(b *testing.B) {
	filter := NewIPv4Filter()
	filter.mode = modeList
	benchmarkFilter(b, filter, 128)
}

func BenchmarkIPv4FilterModeMaps128(b *testing.B) {
	filter := NewIPv4Filter()
	filter.mode = modeMaps
	for i := 0; i < len(filter.ipMaps); i++ {
		filter.ipMaps[i] = make(map[uint32]bool)
	}
	benchmarkFilter(b, filter, 128)
}

func BenchmarkSimpleIPNetList256(b *testing.B) {
	simple := &SimpleIPNetList{}
	benchmarkFilter(b, simple, 256)
}

func BenchmarkIPv4FilterModeList256(b *testing.B) {
	filter := NewIPv4Filter()
	filter.mode = modeList
	benchmarkFilter(b, filter, 256)
}

func BenchmarkIPv4FilterModeMaps256(b *testing.B) {
	filter := NewIPv4Filter()
	filter.mode = modeMaps
	for i := 0; i < len(filter.ipMaps); i++ {
		filter.ipMaps[i] = make(map[uint32]bool)
	}
	benchmarkFilter(b, filter, 256)
}

func BenchmarkSimpleIPNetList512(b *testing.B) {
	simple := &SimpleIPNetList{}
	benchmarkFilter(b, simple, 512)
}

func BenchmarkIPv4FilterModeMaps512(b *testing.B) {
	filter := NewIPv4Filter()
	filter.mode = modeMaps
	for i := 0; i < len(filter.ipMaps); i++ {
		filter.ipMaps[i] = make(map[uint32]bool)
	}
	benchmarkFilter(b, filter, 512)
}