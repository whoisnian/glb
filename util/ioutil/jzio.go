package ioutil

import (
	"compress/flate"
	"encoding/json"
	"io"
)

type Reader struct {
	rd      io.Reader
	jDec    *json.Decoder
	zReader io.ReadCloser
}

func NewReader(r io.Reader) *Reader {
	jz := &Reader{rd: r}
	jz.zReader = flate.NewReader(r)
	jz.jDec = json.NewDecoder(jz.zReader)
	return jz
}

func (jz *Reader) UnMarshal(v any) error {
	if jz.jDec.More() {
		return jz.jDec.Decode(v)
	}
	return io.EOF
}

func (jz *Reader) Close() error {
	return jz.zReader.Close()
}

type Writer struct {
	wr      io.Writer
	jEnc    *json.Encoder
	zWriter *flate.Writer
}

func NewWriter(w io.Writer) (jz *Writer, err error) {
	jz = &Writer{wr: w}
	jz.zWriter, err = flate.NewWriter(w, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}
	jz.jEnc = json.NewEncoder(jz.zWriter)
	return jz, nil
}

func (jz *Writer) Marshal(v any) error {
	if err := jz.jEnc.Encode(v); err != nil {
		return err
	}
	return jz.zWriter.Flush()
}

func (jz *Writer) Close() error {
	return jz.zWriter.Close()
}

type ReadWriter struct {
	*Reader
	*Writer
}

func NewReadWriter(r io.Reader, w io.Writer) (rw *ReadWriter, err error) {
	rw = &ReadWriter{Reader: NewReader(r)}
	if rw.Writer, err = NewWriter(w); err != nil {
		return nil, err
	}
	return rw, nil
}

func (jz *ReadWriter) Close() error {
	if err := jz.Reader.Close(); err != nil {
		return err
	}
	return jz.Writer.Close()
}
