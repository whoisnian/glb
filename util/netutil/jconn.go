package netutil

import (
	"compress/gzip"
	"encoding/json"
	"net"
)

type JConn struct {
	conn    net.Conn
	gWriter *gzip.Writer
	gReader *gzip.Reader
	jEnc    *json.Encoder
	jDec    *json.Decoder
}

func (jconn *JConn) Close() {
	jconn.gReader.Close()
	jconn.gWriter.Close()
}

func (jconn *JConn) Send(v interface{}) {
	jconn.jEnc.Encode(v)
	jconn.gWriter.Flush()
}

func (jconn *JConn) Accept(v interface{}) bool {
	if jconn.jDec.More() {
		jconn.jDec.Decode(v)
		return true
	}
	return false
}

func NewJConn(conn net.Conn) *JConn {
	jconn := &JConn{conn: conn}
	jconn.gWriter = gzip.NewWriter(conn)
	jconn.gWriter.Flush() // send gzip Header immediately
	jconn.jEnc = json.NewEncoder(jconn.gWriter)

	jconn.gReader, _ = gzip.NewReader(conn)
	jconn.jDec = json.NewDecoder(jconn.gReader)
	return jconn
}
