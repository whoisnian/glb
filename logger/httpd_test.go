package logger

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"regexp"
	"testing"

	"github.com/whoisnian/glb/httpd"
)

func requestDiscard(t *testing.T, method string, url string) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("http.NewRequest(%v, %v) got error %v", method, url, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DefaultClient.Do(%v, %v) got error %v", method, url, err)
	}
	resp.Body.Close()
}

func TestRelay4(t *testing.T) {
	var buf bytes.Buffer
	var l *Logger = New(NewTextHandler(&buf, Options{LevelInfo, false, false}))

	mux := httpd.NewMux()
	mux.HandleMiddleware(l.NewMiddleware())
	mux.Handle("/200", http.MethodGet, func(s *httpd.Store) { s.W.WriteHeader(200) })
	mux.Handle("/400", http.MethodPost, func(s *httpd.Store) { s.W.WriteHeader(400) })
	mux.Handle("/403", http.MethodPut, func(s *httpd.Store) { s.W.WriteHeader(403) })
	mux.Handle("/404", http.MethodDelete, func(s *httpd.Store) { s.W.WriteHeader(404) })
	mux.Handle("/empty", http.MethodGet, func(s *httpd.Store) {})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	server := &http.Server{Addr: ln.Addr().String(), Handler: mux}
	go server.Serve(ln)
	defer server.Shutdown(context.Background())

	var tests = []struct {
		method string
		path   string
		code   string
	}{
		{http.MethodGet, "/200", "200"},
		{http.MethodPost, "/400", "400"},
		{http.MethodPut, "/403", "403"},
		{http.MethodDelete, "/404", "404"},
		{http.MethodGet, "/not-found", "404"},
		{http.MethodGet, "/empty", "200"},
	}
	for _, test := range tests {
		buf.Reset()
		requestDiscard(t, test.method, "http://"+server.Addr+test.path)

		reL := `time=` + reTextTime + ` level=INFO msg="" `
		reR := `request.ip=127.0.0.1 request.method=` + test.method + ` request.path=` + test.path + `\n`
		re := `^` + reL + `request.tag=REQ_BEG ` + reR + reL + `request.tag=REQ_END request.code=` + test.code + ` request.dur=[0-9]+ ` + reR + `$`
		if !regexp.MustCompile(re).Match(buf.Bytes()) {
			t.Fatalf("request log should match %q is %q", re, buf.Bytes())
		}
	}
}

func TestRelay6(t *testing.T) {
	var buf bytes.Buffer
	var l *Logger = New(NewTextHandler(&buf, Options{LevelInfo, false, false}))

	mux := httpd.NewMux()
	mux.HandleMiddleware(l.NewMiddleware())
	mux.Handle("/200", http.MethodGet, func(s *httpd.Store) { s.W.WriteHeader(200) })
	mux.Handle("/400", http.MethodPost, func(s *httpd.Store) { s.W.WriteHeader(400) })
	mux.Handle("/403", http.MethodPut, func(s *httpd.Store) { s.W.WriteHeader(403) })
	mux.Handle("/404", http.MethodDelete, func(s *httpd.Store) { s.W.WriteHeader(404) })
	mux.Handle("/empty", http.MethodGet, func(s *httpd.Store) {})

	ln, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	server := &http.Server{Addr: ln.Addr().String(), Handler: mux}
	go server.Serve(ln)
	defer server.Shutdown(context.Background())

	var tests = []struct {
		method string
		path   string
		code   string
	}{
		{http.MethodGet, "/200", "200"},
		{http.MethodPost, "/400", "400"},
		{http.MethodPut, "/403", "403"},
		{http.MethodDelete, "/404", "404"},
		{http.MethodGet, "/not-found", "404"},
		{http.MethodGet, "/empty", "200"},
	}
	for _, test := range tests {
		buf.Reset()
		requestDiscard(t, test.method, "http://"+server.Addr+test.path)

		reL := `time=` + reTextTime + ` level=INFO msg="" `
		reR := `request.ip=::1 request.method=` + test.method + ` request.path=` + test.path + `\n`
		re := `^` + reL + `request.tag=REQ_BEG ` + reR + reL + `request.tag=REQ_END request.code=` + test.code + ` request.dur=[0-9]+ ` + reR + `$`
		if !regexp.MustCompile(re).Match(buf.Bytes()) {
			t.Fatalf("request log should match %q is %q", re, buf.Bytes())
		}
	}
}

func TestRelayRecover(t *testing.T) {
	var buf bytes.Buffer
	var l *Logger = New(NewTextHandler(&buf, Options{LevelInfo, false, false}))

	mux := httpd.NewMux()
	mux.HandleMiddleware(l.NewMiddleware())
	mux.Handle("/panic", http.MethodGet, func(s *httpd.Store) { panic("expected") })

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	server := &http.Server{Addr: ln.Addr().String(), Handler: mux}
	go server.Serve(ln)
	defer server.Shutdown(context.Background())

	requestDiscard(t, http.MethodGet, "http://"+server.Addr+"/panic")

	reL := `time=` + reTextTime + ` level=INFO msg="" `
	reR := `request.ip=127.0.0.1 request.method=` + http.MethodGet + ` request.path=/panic\n`
	re := `^` + reL + `request.tag=REQ_BEG ` + reR +
		`time=` + reTextTime + ` level=ERROR msg="goroutine[^"]+" request.panic=expected\n` +
		reL + `request.tag=REQ_END request.code=500 request.dur=[0-9]+ ` + reR + `$`
	if !regexp.MustCompile(re).Match(buf.Bytes()) {
		t.Fatalf("request log should match %q is %q", re, buf.Bytes())
	}
}
