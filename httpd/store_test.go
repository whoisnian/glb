package httpd_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/whoisnian/glb/httpd"
)

type fakeResponseWriter struct {
	code   int
	buf    bytes.Buffer
	header http.Header
}

func (rw *fakeResponseWriter) Reset()                      { rw.code = 0; rw.buf.Reset(); rw.header = make(http.Header) }
func (rw *fakeResponseWriter) Header() http.Header         { return rw.header }
func (rw *fakeResponseWriter) WriteHeader(statusCode int)  { rw.code = statusCode }
func (rw *fakeResponseWriter) Write(b []byte) (int, error) { return rw.buf.Write(b) }
func (rw *fakeResponseWriter) Flush()                      { rw.code = -1 }

type fakeErrorFlusher struct {
	fakeResponseWriter
}

func (rw *fakeErrorFlusher) FlushError() error { rw.code = -2; return nil }

func TestFlush(t *testing.T) {
	w := &fakeResponseWriter{}
	store := &httpd.Store{W: &httpd.ResponseWriter{Origin: w}}
	store.W.Flush()
	if w.code != -1 {
		t.Fatal("ResponseWriter.Flush should call Origin.Flush successfully")
	}
}

func TestFlushError(t *testing.T) {
	w := &fakeResponseWriter{}
	store := &httpd.Store{W: &httpd.ResponseWriter{Origin: w}}
	err := store.W.FlushError()
	if err != nil || w.code != -1 {
		t.Fatal("ResponseWriter.FlushError should call Origin.Flush successfully")
	}

	f := &fakeErrorFlusher{}
	store = &httpd.Store{W: &httpd.ResponseWriter{Origin: f}}
	err = store.W.FlushError()
	if err != nil || f.code != -2 {
		t.Fatal("ResponseWriter.FlushError should call Origin.FlushError successfully")
	}
}

func TestCreateHandler(t *testing.T) {
	store := &httpd.Store{W: &httpd.ResponseWriter{}, R: &http.Request{}}
	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		if w != store.W || r != store.R {
			t.Fatal("CreateHandler should pass original Request and ResponseWriter to httpHandler")
		}
	}
	httpd.CreateHandler(httpHandler)(store)
}

func TestGetClientIP(t *testing.T) {
	store := &httpd.Store{W: &httpd.ResponseWriter{}, R: &http.Request{Header: make(http.Header)}}
	if got := store.GetClientIP(); got != "" {
		t.Fatalf("store.GetClientIP() = %q, want %q", got, "")
	}
	// Request.RemoteAddr
	store.R.RemoteAddr = "127.0.0.1"
	if got := store.GetClientIP(); got != "127.0.0.1" {
		t.Fatalf("store.GetClientIP() = %q, want %q", got, "127.0.0.1")
	}
	// Header["X-Real-IP"]
	store.R.Header.Set("X-Real-IP", "172.31.0.10")
	if got := store.GetClientIP(); got != "172.31.0.10" {
		t.Fatalf("store.GetClientIP() = %q, want %q", got, "172.31.0.10")
	}
	// Header["X-Forwarded-For"]
	store.R.Header.Set("X-Forwarded-For", "192.168.0.2")
	if got := store.GetClientIP(); got != "192.168.0.2" {
		t.Fatalf("store.GetClientIP() = %q, want %q", got, "192.168.0.2")
	}
	// Header["X-Forwarded-For"]
	store.R.Header.Set("X-Forwarded-For", "192.168.0.3, 172.31.0.10")
	if got := store.GetClientIP(); got != "192.168.0.3" {
		t.Fatalf("store.GetClientIP() = %q, want %q", got, "192.168.0.3")
	}
	// Header["X-Client-IP"]
	store.R.Header.Set("X-Client-IP", "8.8.8.8")
	if got := store.GetClientIP(); got != "8.8.8.8" {
		t.Fatalf("store.GetClientIP() = %q, want %q", got, "8.8.8.8")
	}
}

func TestCookieValue(t *testing.T) {
	tests := []struct {
		k, v string
	}{
		{"value1", "hello"},
		{"value2", "123"},
		{"value3", "false"},
		{"value4", ""},
	}

	cookieStr := ""
	for _, tt := range tests {
		if tt.v != "" {
			cookieStr = cookieStr + tt.k + "=" + tt.v + ";"
		}
	}
	store := &httpd.Store{W: &httpd.ResponseWriter{}, R: &http.Request{
		Header: http.Header{"Cookie": {cookieStr}},
	}}

	for _, tt := range tests {
		if res := store.CookieValue(tt.k); res != tt.v {
			t.Fatalf("CookieValue(%q) = %q, want %q", tt.k, res, tt.v)
		}
	}
}

func TestRespond200(t *testing.T) {
	w := &fakeResponseWriter{}
	store := &httpd.Store{W: &httpd.ResponseWriter{Origin: w}}
	store.Respond200(nil)
	if w.code != http.StatusOK || w.buf.Len() != 0 {
		t.Fatalf("Respond200(nil) = %d %q, want %d nil", w.code, w.buf.Bytes(), http.StatusOK)
	}

	w.Reset()
	data := []byte("hello")
	store.Respond200(data)
	if w.code != http.StatusOK || !bytes.Equal(data, w.buf.Bytes()) {
		t.Fatalf("Respond200(data) = %d %q, want %d %q", w.code, w.buf.Bytes(), http.StatusOK, data)
	}
}

func TestRespondJson(t *testing.T) {
	w := &fakeResponseWriter{header: make(http.Header)}
	store := &httpd.Store{W: &httpd.ResponseWriter{Origin: w}}

	type jsonTest struct {
		A int
		B float64
		C string
		D bool
		E []int
		F []byte
		G [][]byte
	}

	input := jsonTest{0, 0.5, "hello", true, []int{-1, 1}, []byte("array"), [][]byte{[]byte("null"), nil}}
	want := `{"A":0,"B":0.5,"C":"hello","D":true,"E":[-1,1],"F":"YXJyYXk=","G":["bnVsbA==",null]}`
	store.RespondJson(http.StatusOK, input)
	if w.code != http.StatusOK || w.Header().Get("Content-Type") != "application/json; charset=utf-8" || w.buf.String() != want+"\n" {
		t.Fatalf("RespondJson(input) = %d %s, want %d %s", w.code, w.buf.String(), http.StatusOK, want)
	}
}

func TestRedirect(t *testing.T) {
	w := &fakeResponseWriter{header: make(http.Header)}
	store := &httpd.Store{W: &httpd.ResponseWriter{Origin: w}, R: &http.Request{}}

	url := "http://127.0.0.1:8000/redirect"
	store.Redirect(http.StatusFound, url)
	if w.code != http.StatusFound || w.Header().Get("Location") != url {
		t.Fatalf("Redirect(url, code) = %d %s, want %d %s", w.code, w.Header().Get("Location"), http.StatusFound, url)
	}
}

func TestError404(t *testing.T) {
	w := &fakeResponseWriter{header: make(http.Header)}
	store := &httpd.Store{W: &httpd.ResponseWriter{Origin: w}}

	msg := "TestError404"
	store.Error404(msg)
	if w.code != http.StatusNotFound || w.buf.String() != msg+"\n" {
		t.Fatalf("Error404(msg) = %d %s, want %d %s", w.code, w.buf.String(), http.StatusNotFound, msg)
	}
}

func TestError500(t *testing.T) {
	w := &fakeResponseWriter{header: make(http.Header)}
	store := &httpd.Store{W: &httpd.ResponseWriter{Origin: w}}

	msg := "TestError500"
	store.Error500(msg)
	if w.code != http.StatusInternalServerError || w.buf.String() != msg+"\n" {
		t.Fatalf("Error500(msg) = %d %s, want %d %s", w.code, w.buf.String(), http.StatusNotFound, msg)
	}
}
