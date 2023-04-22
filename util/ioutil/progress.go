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

func NewProgressWriter(w io.Writer) *ProgressWriter {
	return &ProgressWriter{
		wr:     w,
		size:   0,
		status: nil,
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

func (pw *ProgressWriter) Size() int {
	return pw.size
}

func (pw *ProgressWriter) Status() chan int {
	if pw.status == nil {
		pw.status = make(chan int)
	}
	return pw.status
}

func (pw *ProgressWriter) Close() {
	if pw.status != nil {
		pw.status <- pw.size
		close(pw.status)
	}
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.wr.Write(p)
	pw.sum(n)
	return n, err
}

func (pw *ProgressWriter) WriteString(s string) (n int, err error) {
	if sw, ok := pw.wr.(io.StringWriter); ok {
		n, err = sw.WriteString(s)
	} else {
		n, err = pw.wr.Write(strutil.UnsafeStringToBytes(s))
	}
	pw.sum(n)
	return n, err
}
