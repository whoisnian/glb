package ioutil

import (
	"io"

	"github.com/whoisnian/glb/util/strutil"
)

type ProgressWriter struct {
	wr     io.Writer
	size   int64
	status chan int64
}

// NewProgressWriter wraps io.Writer with total written size and a status channel.
func NewProgressWriter(w io.Writer) *ProgressWriter {
	return &ProgressWriter{
		wr:     w,
		size:   0,
		status: make(chan int64),
	}
}

func (pw *ProgressWriter) sum(n int) {
	pw.size += int64(n)
	select {
	case pw.status <- pw.size:
	default:
	}
}

// Size returns total written size of io.Writer.
func (pw *ProgressWriter) Size() int64 {
	return pw.size
}

// Status returns the status channel.
// Total written size will be sent to the status channel without blocking after every write operation.
func (pw *ProgressWriter) Status() chan int64 {
	return pw.status
}

// Close closes the status channel.
// It should be executed only by the sender, never the receiver.
func (pw *ProgressWriter) Close() {
	close(pw.status)
}

// Write implements the standard io.Writer interface.
func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.wr.Write(p)
	if n > 0 {
		pw.sum(n)
	}
	return n, err
}

// WriteString implements the standard io.StringWriter interface.
func (pw *ProgressWriter) WriteString(s string) (n int, err error) {
	if sw, ok := pw.wr.(io.StringWriter); ok {
		n, err = sw.WriteString(s)
	} else {
		n, err = pw.wr.Write(strutil.UnsafeStringToBytes(s))
	}
	if n > 0 {
		pw.sum(n)
	}
	return n, err
}
