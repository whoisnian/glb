package tasklane

import (
	"crypto/rand"
	"strconv"
	"sync"
	"sync/atomic"
)

const (
	base32alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	threshold      = 1 << 63
)

// Sequence generates incremental sequence numbers with a random prefix.
// When the sequence number exceeds 2^63, it will try to reset the prefix and sequence number.
type Sequence struct {
	prefixLen int
	prefixVal atomic.Value

	seq atomic.Uint64
	mu  sync.Mutex
}

// NewSequence creates a new sequence with the given prefix length.
func NewSequence(prefixLen int) *Sequence {
	s := &Sequence{
		prefixLen: prefixLen,
	}
	s.reset()
	return s
}

func (s *Sequence) reset() {
	prefix := make([]byte, s.prefixLen, s.prefixLen+1)
	rand.Read(prefix) // crypto/rand.Read() never returns an error
	for i := range prefix {
		prefix[i] = base32alphabet[prefix[i]%32]
	}
	s.prefixVal.Store(append(prefix, '-'))
	s.seq.Store(0)
}

// Next returns the next prefixed sequence number.
// It is safe to call Next from multiple goroutines.
func (s *Sequence) Next() []byte {
	if s.seq.Load() > threshold {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.seq.Load() > threshold {
			s.reset()
		}
	}
	next := s.seq.Add(1) // sequence number may exceed 2^63, but it's ok for uint64
	return strconv.AppendUint(s.prefixVal.Load().([]byte), next, 16)
}
