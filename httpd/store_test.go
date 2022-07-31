package httpd

import (
	"bytes"
	"net/http"
	"testing"
)

type fakeResponseWriter struct {
	code   int
	buf    bytes.Buffer
	header http.Header
}

func (rw *fakeResponseWriter) Reset()                     { rw.code = 0; rw.buf.Reset(); rw.header = make(http.Header) }
func (rw *fakeResponseWriter) Header() http.Header        { return rw.header }
func (rw *fakeResponseWriter) WriteHeader(statusCode int) { rw.code = statusCode }
func (rw *fakeResponseWriter) Write(b []byte) (int, error) {
	if rw.code == 0 {
		rw.code = http.StatusOK
	}
	return rw.buf.Write(b)
}

func TestCreateHandler(t *testing.T) {
	store := &Store{W: &fakeResponseWriter{}, R: &http.Request{}}
	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		if w != store.W || r != store.R {
			t.Fatal("CreateHandler should pass original Request and ResponseWriter to httpHandler")
		}
	}
	CreateHandler(httpHandler)(store)
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
	store := &Store{W: &fakeResponseWriter{}, R: &http.Request{
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
	store := &Store{W: w}
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
	store := &Store{W: w}

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
	store.RespondJson(input)
	if w.code != http.StatusOK || w.Header().Get("Content-Type") != "application/json; charset=utf-8" || w.buf.String() != want+"\n" {
		t.Fatalf("RespondJson(input) = %d %s, want %d %s", w.code, w.buf.String(), http.StatusOK, want)
	}
}

func TestRedirect(t *testing.T) {
	w := &fakeResponseWriter{header: make(http.Header)}
	store := &Store{W: w, R: &http.Request{}}

	url := "http://127.0.0.1:8000/redirect"
	store.Redirect(url, http.StatusFound)
	if w.code != http.StatusFound || w.Header().Get("Location") != url {
		t.Fatalf("Redirect(url, code) = %d %s, want %d %s", w.code, w.Header().Get("Location"), http.StatusFound, url)
	}
}

func TestError404(t *testing.T) {
	w := &fakeResponseWriter{header: make(http.Header)}
	store := &Store{W: w}

	msg := "TestError404"
	store.Error404(msg)
	if w.code != http.StatusNotFound || w.buf.String() != msg+"\n" {
		t.Fatalf("Error404(msg) = %d %s, want %d %s", w.code, w.buf.String(), http.StatusNotFound, msg)
	}
}

func TestError500(t *testing.T) {
	w := &fakeResponseWriter{header: make(http.Header)}
	store := &Store{W: w}

	msg := "TestError500"
	store.Error500(msg)
	if w.code != http.StatusInternalServerError || w.buf.String() != msg+"\n" {
		t.Fatalf("Error500(msg) = %d %s, want %d %s", w.code, w.buf.String(), http.StatusNotFound, msg)
	}
}
