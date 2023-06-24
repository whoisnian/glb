package ioutil

import (
	"io"

	"github.com/whoisnian/glb/util/strutil"
)

type ProgressWriter struct {
	wr     io.Writer
	size   int
	status chan int
}

// NewProgressWriter wraps io.Writer with total written size and a status channel.
func NewProgressWriter(w io.Writer) *ProgressWriter {
	return &ProgressWriter{
		wr:     w,
		size:   0,
		status: make(chan int),
	}
}

func (pw *ProgressWriter) sum(n int) {
	pw.size += n
	if pw.status != nil {
		select {
		case pw.status <- pw.size:
		default:
		}
	}
}

// Size returns total written size of io.Writer.
func (pw *ProgressWriter) Size() int {
	return pw.size
}

// Status returns the status channel.
// Total written size will be sent to the status channel without blocking after every write operation.
func (pw *ProgressWriter) Status() chan int {
	return pw.status
}

// Close sends total written size to the blocking channel and then closes the channel.
// Only the sender should close a channel, never the receiver.
func (pw *ProgressWriter) Close() {
	if pw.status != nil {
		pw.status <- pw.size
		close(pw.status)
	}
}

// Write implements the standard io.Writer interface.
func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.wr.Write(p)
	pw.sum(n)
	return n, err
}

// WriteString implements the standard io.StringWriter interface.
func (pw *ProgressWriter) WriteString(s string) (n int, err error) {
	if sw, ok := pw.wr.(io.StringWriter); ok {
		n, err = sw.WriteString(s)
	} else {
		n, err = pw.wr.Write(strutil.UnsafeStringToBytes(s))
	}
	pw.sum(n)
	return n, err
}
