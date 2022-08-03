package netutil

import (
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

var ErrInvalidIPv4CIDR = errors.New("invalid IPv4 CIDR")

var ipv4Masks = [...]uint32{
	0x80000000, 0xc0000000, 0xe0000000, 0xf0000000,
	0xf8000000, 0xfc000000, 0xfe000000, 0xff000000,
	0xff800000, 0xffc00000, 0xffe00000, 0xfff00000,
	0xfff80000, 0xfffc0000, 0xfffe0000, 0xffff0000,
	0xffff8000, 0xffffc000, 0xffffe000, 0xfffff000,
	0xfffff800, 0xfffffc00, 0xfffffe00, 0xffffff00,
	0xffffff80, 0xffffffc0, 0xffffffe0, 0xfffffff0,
	0xfffffff8, 0xfffffffc, 0xfffffffe, 0xffffffff,
}

const listSize = 256
const (
	modeList uint32 = iota
	modeMaps
)

type IPv4Filter struct {
	mutex    sync.RWMutex
	matchAll *atomic.Bool

	mode   uint32
	index  int
	ipList [listSize][2]uint32
	ipMaps [32]map[uint32]bool
}

func NewIPv4Filter() *IPv4Filter {
	return &IPv4Filter{matchAll: &atomic.Bool{}, mode: modeList, index: 0}
}

func (f *IPv4Filter) Add(cidr *net.IPNet) error {
	ones, bits := cidr.Mask.Size()
	if bits != 32 || ones > 32 || len(cidr.IP) != net.IPv4len {
		return ErrInvalidIPv4CIDR
	} else if ones == 0 {
		f.matchAll.Store(true)
		return nil
	}

	nip := binary.BigEndian.Uint32(cidr.IP)

	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.mode == modeList {
		if f.index < listSize {
			f.ipList[f.index] = [2]uint32{nip & ipv4Masks[ones-1], uint32(ones)}
			f.index++
		} else {
			f.mode = modeMaps
			for i := 0; i < len(f.ipMaps); i++ {
				f.ipMaps[i] = make(map[uint32]bool)
			}
			for i := 0; i < f.index; i++ {
				if f.ipList[i][1] > 0 {
					f.ipMaps[f.ipList[i][1]-1][f.ipList[i][0]] = true
				}
			}
			f.ipMaps[ones-1][nip&ipv4Masks[ones-1]] = true
		}
	} else {
		f.ipMaps[ones-1][nip&ipv4Masks[ones-1]] = true
	}
	return nil
}

func (f *IPv4Filter) Remove(cidr *net.IPNet) error {
	ones, bits := cidr.Mask.Size()
	if bits != 32 || ones > 32 || len(cidr.IP) != net.IPv4len {
		return ErrInvalidIPv4CIDR
	} else if ones == 0 {
		f.matchAll.Store(false)
		return nil
	}

	nip := binary.BigEndian.Uint32(cidr.IP)

	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.mode == modeList {
		for i := 0; i < f.index; i++ {
			if uint32(ones) == f.ipList[i][1] && nip&ipv4Masks[ones-1] == f.ipList[i][0] {
				f.ipList[i] = [2]uint32{0, 0} // reset to invalid CIDR
			}
		}
	} else {
		delete(f.ipMaps[ones-1], nip&ipv4Masks[ones-1])
	}
	return nil
}

func (f *IPv4Filter) Contains(ip net.IP) bool {
	if f.matchAll.Load() {
		return true
	} else if len(ip) != net.IPv4len {
		return false
	}

	nip := binary.BigEndian.Uint32(ip)

	f.mutex.RLock()
	defer f.mutex.RUnlock()
	if f.mode == modeList {
		for i := 0; i < f.index; i++ {
			if f.ipList[i][1] > 0 && nip&ipv4Masks[f.ipList[i][1]-1] == f.ipList[i][0] {
				return true
			}
		}
	} else {
		for i := 0; i < 32; i++ {
			if f.ipMaps[i][nip&ipv4Masks[i]] {
				return true
			}
		}
	}
	return false
}
