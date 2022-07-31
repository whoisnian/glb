// Package jzio implements encoding and decoding of compressed JSON stream.
package jzio

import (
	"compress/flate"
	"encoding/json"
	"io"
)

// Decoder reads and decodes JSON values from an input stream.
type Decoder struct {
	rd      io.Reader
	jDec    *json.Decoder
	zReader io.ReadCloser
}

// NewDecoder returns a new decoder that reads from r.
//
// Like json.Decoder, the decoder may read data from r beyond the JSON values requested.
func NewDecoder(r io.Reader) *Decoder {
	jz := &Decoder{rd: r}
	jz.zReader = flate.NewReader(r)
	jz.jDec = json.NewDecoder(jz.zReader)
	return jz
}

// UnMarshal reads next JSON value and stores it in the value pointed to by v.
func (jz *Decoder) UnMarshal(v any) error {
	if jz.jDec.More() {
		return jz.jDec.Decode(v)
	}
	return io.EOF
}

// Close closes the decoder and returns error during decompression.
func (jz *Decoder) Close() error {
	return jz.zReader.Close()
}

// Encoder writes JSON values to an output stream.
type Encoder struct {
	wr      io.Writer
	jEnc    *json.Encoder
	zWriter *flate.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) (jz *Encoder, err error) {
	jz = &Encoder{wr: w}
	jz.zWriter, err = flate.NewWriter(w, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}
	jz.jEnc = json.NewEncoder(jz.zWriter)
	return jz, nil
}

// Marshal writes the JSON encoding of v to the stream, followed by a newline.
func (jz *Encoder) Marshal(v any) error {
	if err := jz.jEnc.Encode(v); err != nil {
		return err
	}
	return jz.zWriter.Flush()
}

// Close flushes and closes the encoder.
func (jz *Encoder) Close() error {
	return jz.zWriter.Close()
}

// Codec groups the jzio decoder and encoder.
type Codec struct {
	*Decoder
	*Encoder
}

// NewCodec returns a new codec that reads from r and writes to w.
func NewCodec(r io.Reader, w io.Writer) (rw *Codec, err error) {
	rw = &Codec{Decoder: NewDecoder(r)}
	if rw.Encoder, err = NewEncoder(w); err != nil {
		return nil, err
	}
	return rw, nil
}

// Close closes the decoder and then the encoder.
func (jz *Codec) Close() error {
	if err := jz.Decoder.Close(); err != nil {
		return err
	}
	return jz.Encoder.Close()
}
