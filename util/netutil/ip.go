package netutil

import (
	"errors"
	"net"
	"net/netip"
	"sync"
)

var (
	ErrInvalidIPv4CIDR = errors.New("invalid IPv4 CIDR")

	IPv4Mask [32]net.IPMask
)

func init() {
	for i := 0; i < 32; i++ {
		IPv4Mask[i] = net.CIDRMask(i+1, 32)
	}
}

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

type Checklist4 struct {
	mutex    sync.RWMutex
	matchAll bool
	ipMaps   [32]map[netip.Addr]bool
}

func NewWhitelist4() *Checklist4 {
	w := &Checklist4{}
	for i := 0; i < 32; i++ {
		w.ipMaps[i] = make(map[netip.Addr]bool)
	}
	return w
}

func (w *Checklist4) Add(cidr *net.IPNet) error {
	ones, bits := cidr.Mask.Size()
	if bits != 32 || ones > 32 || len(cidr.IP) != net.IPv4len {
		return ErrInvalidIPv4CIDR
	} else if ones == 0 {
		w.matchAll = true
		return nil
	}

	nip := netip.AddrFrom4(*(*[4]byte)(cidr.IP.Mask(cidr.Mask)))
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.ipMaps[ones-1][nip] = true
	return nil
}

func (w *Checklist4) Remove(cidr *net.IPNet) error {
	ones, bits := cidr.Mask.Size()
	if bits != 32 || ones > 32 || len(cidr.IP) != net.IPv4len {
		return ErrInvalidIPv4CIDR
	} else if ones == 0 {
		w.matchAll = false
		return nil
	}

	nip := netip.AddrFrom4(*(*[4]byte)(cidr.IP.Mask(cidr.Mask)))
	w.mutex.Lock()
	defer w.mutex.Unlock()
	delete(w.ipMaps[ones-1], nip)
	return nil
}

func (w *Checklist4) Contains(ip net.IP) bool {
	if w.matchAll {
		return true
	} else if len(ip) != net.IPv4len {
		return false
	}

	w.mutex.RLock()
	defer w.mutex.RUnlock()
	for i := 0; i < 32; i++ {
		nip := netip.AddrFrom4(*(*[4]byte)(ip.Mask(IPv4Mask[i])))
		if w.ipMaps[i][nip] {
			return true
		}
	}
	return false
}
