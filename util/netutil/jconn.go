package netutil

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net"
)

type JConn struct {
	conn    net.Conn
	gWriter *gzip.Writer
	gReader *gzip.Reader
	jEnc    *json.Encoder
	jDec    *json.Decoder
}

// Close only close gzip.Reader/gzip.Writer. Source net.Conn should be closed manually.
func (jconn *JConn) Close() (err error) {
	if err = jconn.gReader.Close(); err != nil {
		return err
	}
	return jconn.gWriter.Close()
}

func (jconn *JConn) Send(v interface{}) (err error) {
	if err = jconn.jEnc.Encode(v); err != nil {
		return err
	}
	return jconn.gWriter.Flush()
}

func (jconn *JConn) Accept(v interface{}) (err error) {
	if jconn.jDec.More() {
		return jconn.jDec.Decode(v)
	}
	return io.EOF
}

func NewJConn(conn net.Conn) (jconn *JConn, err error) {
	jconn = &JConn{conn: conn}
	jconn.gWriter = gzip.NewWriter(conn)
	if err = jconn.gWriter.Flush(); err != nil {
		return nil, err
	}
	jconn.jEnc = json.NewEncoder(jconn.gWriter)

	if jconn.gReader, err = gzip.NewReader(conn); err != nil {
		return nil, err
	}
	jconn.jDec = json.NewDecoder(jconn.gReader)
	return jconn, err
}
