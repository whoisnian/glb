package logger

import (
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 1024)
		return &buf
	},
}

func newBuffer() *[]byte {
	return bufferPool.Get().(*[]byte)
}

func freeBuffer(buf *[]byte) {
	const maxBufferSize = 16 << 10
	if cap(*buf) <= maxBufferSize {
		*buf = (*buf)[:0]
		bufferPool.Put(buf)
	}
}

const smallsString = "00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

// 0 <= i <= 9
// func appendIntWidth1(buf *[]byte, i int) {
// 	*buf = append(*buf, smallsString[i*2+1])
// }

// 00 <= i <= 99
func appendIntWidth2(buf *[]byte, i int) {
	*buf = append(*buf, smallsString[i*2:i*2+2]...)
}

// 000 <= i <= 999
// func appendIntWidth3(buf *[]byte, i int) {
// 	l := i / 100
// 	i -= l * 100
// 	*buf = append(*buf, smallsString[l*2+1])
// 	*buf = append(*buf, smallsString[i*2:i*2+2]...)
// }

// 0000 <= i <= 9999
func appendIntWidth4(buf *[]byte, i int) {
	l := i / 100
	i -= l * 100
	*buf = append(*buf, smallsString[l*2:l*2+2]...)
	*buf = append(*buf, smallsString[i*2:i*2+2]...)
}
