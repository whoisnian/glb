package netutil

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

var (
	ErrInvalidIPv4CIDR = errors.New("invalid IPv4 CIDR")
)

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
	matchAll uint32
	uintMaps [32]map[uint32]bool
}

func NewChecklist4() *Checklist4 {
	w := &Checklist4{}
	for i := 0; i < 32; i++ {
		w.uintMaps[i] = make(map[uint32]bool)
	}
	return w
}

func (w *Checklist4) Add(cidr *net.IPNet) error {
	ones, bits := cidr.Mask.Size()
	if bits != 32 || ones > 32 || len(cidr.IP) != net.IPv4len {
		return ErrInvalidIPv4CIDR
	} else if ones == 0 {
		atomic.StoreUint32(&w.matchAll, 1)
		return nil
	}

	nip := uint32(cidr.IP[0]&cidr.Mask[0]) |
		uint32(cidr.IP[1]&cidr.Mask[1])<<8 |
		uint32(cidr.IP[2]&cidr.Mask[2])<<16 |
		uint32(cidr.IP[3]&cidr.Mask[3])<<24
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.uintMaps[ones-1][nip] = true
	return nil
}

func (w *Checklist4) Remove(cidr *net.IPNet) error {
	ones, bits := cidr.Mask.Size()
	if bits != 32 || ones > 32 || len(cidr.IP) != net.IPv4len {
		return ErrInvalidIPv4CIDR
	} else if ones == 0 {
		atomic.StoreUint32(&w.matchAll, 0)
		return nil
	}

	nip := uint32(cidr.IP[0]&cidr.Mask[0]) |
		uint32(cidr.IP[1]&cidr.Mask[1])<<8 |
		uint32(cidr.IP[2]&cidr.Mask[2])<<16 |
		uint32(cidr.IP[3]&cidr.Mask[3])<<24
	w.mutex.Lock()
	defer w.mutex.Unlock()
	delete(w.uintMaps[ones-1], nip)
	return nil
}

func (w *Checklist4) Contains(ip net.IP) bool {
	if atomic.LoadUint32(&w.matchAll) == 1 {
		return true
	} else if len(ip) != net.IPv4len {
		return false
	}

	nip := uint32(0)
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	for i := 0; i < 32; i++ {
		nip |= uint32(ip[i>>3]&(0x80>>(i&7))) << (i & 0xF8)
		if w.uintMaps[i][nip] {
			return true
		}
	}
	return false
}
