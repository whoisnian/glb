package ioutil

import (
	"compress/gzip" // TODO: test stateless gzip
	"encoding/json"
	"io"
)

type Reader struct {
	rd      io.Reader
	jDec    *json.Decoder
	zReader *gzip.Reader
}

func NewReader(r io.Reader) (jz *Reader, err error) {
	jz = &Reader{rd: r}
	if jz.zReader, err = gzip.NewReader(r); err != nil {
		return nil, err
	}
	jz.jDec = json.NewDecoder(jz.zReader)
	return jz, nil
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
	zWriter *gzip.Writer
}

func NewWriter(w io.Writer) (jz *Writer, err error) {
	jz = &Writer{wr: w}
	jz.zWriter = gzip.NewWriter(w)
	if err = jz.zWriter.Flush(); err != nil {
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
	rw = &ReadWriter{}
	if rw.Reader, err = NewReader(r); err != nil {
		return nil, err
	}
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
