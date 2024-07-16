package logger

import (
	"fmt"
	"testing"
)

func TestBufferPoolRace(t *testing.T) {
	const P = 10
	const N = 10000
	done := make(chan struct{})
	for i := 0; i < P; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < N; j++ {
				buf := newBuffer()
				*buf = append(*buf, make([]byte, j)...) // len(*buf) should be j
				for k := 0; k < j; k++ {
					(*buf)[k] = byte(j % 256)
				}
				for k := 0; k < len(*buf); k++ {
					if (*buf)[k] != byte(j%256) {
						t.Errorf("goroutine(%d.%d) read buf[%d] = %d, want %d", i, j, k, (*buf)[k], j%256)
						return
					}
				}
				freeBuffer(buf)
			}
		}()
	}
	for i := 0; i < P; i++ {
		<-done
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
