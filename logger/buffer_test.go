package logger

import (
	"crypto/md5"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/whoisnian/glb/util/ioutil"
)

func TestBufferPool(t *testing.T) {
	var SEED = uint64(time.Now().Unix())
	t.Logf("Running tests with PCG seed (0,%v)", SEED)
	rd := rand.New(rand.NewPCG(0, SEED))

	var (
		tests   = 1 << 10
		size    int
		lastSum [16]byte
		lastCap = initBufferSize
	)
	for i := 0; i < tests; i++ {
		buf := newBuffer()
		// Any item stored in the Pool may be removed automatically at any time without notification.
		if cap(*buf) > initBufferSize {
			if cap(*buf) != lastCap {
				t.Fatalf("cap(newBuffer) = %d, want %d", cap(*buf), lastCap)
			}
			if lastSum != md5.Sum((*buf)[:size]) {
				t.Fatalf("md5(newBuffer) = %q, want %q", md5.Sum((*buf)[:size]), lastSum)
			}
		}
		*buf = append(*buf, make([]byte, rd.IntN(maxBufferSize*1.5))...)
		size, _ = ioutil.ReadRand(rd, *buf)
		lastSum = md5.Sum((*buf)[:size])
		lastCap = cap(*buf)
		freeBuffer(buf)
	}
}

func TestAppendIntWidth1(t *testing.T) {
	buf := make([]byte, 8)
	for i := 0; i <= 9; i++ {
		buf = buf[:0]
		appendIntWidth1(&buf, i)

		got, want := string(buf), fmt.Sprintf("%01d", i)
		if got != want {
			t.Fatalf("appendIntWidth1(%d) = %s, want %s", i, got, want)
		}
	}
}

func TestAppendIntWidth2(t *testing.T) {
	buf := make([]byte, 8)
	for i := 0; i <= 99; i++ {
		buf = buf[:0]
		appendIntWidth2(&buf, i)

		got, want := string(buf), fmt.Sprintf("%02d", i)
		if got != want {
			t.Fatalf("appendIntWidth2(%d) = %s, want %s", i, got, want)
		}
	}
}

func TestAppendIntWidth3(t *testing.T) {
	buf := make([]byte, 8)
	for i := 0; i <= 999; i++ {
		buf = buf[:0]
		appendIntWidth3(&buf, i)

		got, want := string(buf), fmt.Sprintf("%03d", i)
		if got != want {
			t.Fatalf("appendIntWidth3(%d) = %s, want %s", i, got, want)
		}
	}
}

func TestAppendIntWidth4(t *testing.T) {
	buf := make([]byte, 8)
	for i := 0; i <= 9999; i++ {
		buf = buf[:0]
		appendIntWidth4(&buf, i)

		got, want := string(buf), fmt.Sprintf("%04d", i)
		if got != want {
			t.Fatalf("appendIntWidth4(%d) = %s, want %s", i, got, want)
		}
	}
}
