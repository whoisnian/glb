package logger

import (
	"fmt"
	"testing"
)

func TestBufferPoolRace(t *testing.T) {
	const P = 10
	const N = 10000
	done := make(chan struct{})
	for i := range P {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := range N {
				buf := newBuffer()
				*buf = append(*buf, make([]byte, j)...) // len(*buf) should be j
				for k := range j {
					(*buf)[k] = byte(j % 256)
				}
				for k := range *buf {
					if (*buf)[k] != byte(j%256) {
						t.Errorf("goroutine(%d.%d) read buf[%d] = %d, want %d", i, j, k, (*buf)[k], j%256)
						return
					}
				}
				freeBuffer(buf)
			}
		}()
	}
	for range P {
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

func BenchmarkAppendIntWidth1(b *testing.B) {
	buf, input := make([]byte, 8), 0
	b.ReportAllocs()
	for b.Loop() {
		buf = buf[:0]
		appendIntWidth1(&buf, input)
		input += 1
		if input == 10 {
			input = 0
		}
	}
}
func BenchmarkAppendIntWidth2(b *testing.B) {
	buf, input := make([]byte, 8), 0
	b.ReportAllocs()
	for b.Loop() {
		buf = buf[:0]
		appendIntWidth2(&buf, input)
		input += 1
		if input == 100 {
			input = 0
		}
	}
}

func BenchmarkAppendIntWidth3(b *testing.B) {
	buf, input := make([]byte, 8), 0
	b.ReportAllocs()
	for b.Loop() {
		buf = buf[:0]
		appendIntWidth3(&buf, input)
		input += 1
		if input == 1000 {
			input = 0
		}
	}
}

func BenchmarkAppendIntWidth4(b *testing.B) {
	buf, input := make([]byte, 8), 0
	b.ReportAllocs()
	for b.Loop() {
		buf = buf[:0]
		appendIntWidth4(&buf, input)
		input += 1
		if input == 10000 {
			input = 0
		}
	}
}
