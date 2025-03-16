package tasklane

import (
	"bytes"
	"strconv"
	"sync"
	"testing"
)

func TestSequenceNext(t *testing.T) {
	const prefixLen = 8
	s := NewSequence(prefixLen)
	prev := s.Next()
	for range 100 {
		curr := s.Next()
		if !bytes.Equal(prev[:prefixLen], curr[:prefixLen]) {
			t.Errorf("prefix changed from %s to %s", prev, curr)
		}
		prevNum, _ := strconv.ParseUint(string(prev[prefixLen+1:]), 16, 64)
		currNum, _ := strconv.ParseUint(string(curr[prefixLen+1:]), 16, 64)
		if currNum <= prevNum {
			t.Errorf("sequence did not increment: %x(%s) -> %x(%s)", prevNum, prev, currNum, curr)
		}
		prev = curr
	}
}

func TestSequenceReset(t *testing.T) {
	const prefixLen = 8
	s := NewSequence(prefixLen)
	s.seq.Store(threshold - 1)

	x1, x2 := s.Next(), s.Next()
	// reset prefix and sequence number
	y1, y2 := s.Next(), s.Next()

	if !bytes.Equal(x1[:prefixLen], x2[:prefixLen]) {
		t.Errorf("prefix changed from %s to %s", x1, x2)
	}
	if !bytes.Equal(y1[:prefixLen], y2[:prefixLen]) {
		t.Errorf("prefix changed from %s to %s", y1, y2)
	}
	if bytes.Equal(x2[:prefixLen], y1[:prefixLen]) {
		t.Errorf("prefix not reset from %s to %s", x2, y1)
	}
}

func TestSequenceNextRace(t *testing.T) {
	const prefixLen = 8
	s := NewSequence(prefixLen)
	seen := new(sync.Map)

	const P = 100
	const N = 10000
	done := make(chan struct{})
	for range P {
		go func() {
			defer func() { done <- struct{}{} }()
			for range N {
				key := string(s.Next())
				if _, loaded := seen.LoadOrStore(key, true); loaded {
					t.Errorf("duplicate sequence number %s", key)
				}
			}
		}()
	}
	for range P {
		<-done
	}
}
