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

const reTid = `[2-7A-Z]{8}-[0-9a-z]{0,13}`

func requestDiscard(t *testing.T, method string, url string) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("http.NewRequest(%v, %v) get err %v", method, url, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DefaultClient.Do(%v, %v) get err %v", method, url, err)
	}
	resp.Body.Close()
}

func TestRelay4(t *testing.T) {
	var buf bytes.Buffer
	var l *Logger = New(NewTextHandler(&buf, NewOptions(LevelInfo, false, false)))

	mux := httpd.NewMux()
	mux.HandleRelay(l.Relay)
	mux.Handle("/200", http.MethodGet, func(s *httpd.Store) { s.W.WriteHeader(200) })
	mux.Handle("/400", http.MethodPost, func(s *httpd.Store) { s.W.WriteHeader(400) })
	mux.Handle("/403", http.MethodPut, func(s *httpd.Store) { s.W.WriteHeader(403) })
	mux.Handle("/404", http.MethodDelete, func(s *httpd.Store) { s.W.WriteHeader(404) })

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
	}
	for _, test := range tests {
		buf.Reset()
		requestDiscard(t, test.method, "http://"+server.Addr+test.path)

		l := `time=` + reTextTime + ` level=INFO msg="" `
		r := `ip=127.0.0.1 method=` + test.method + ` path=` + test.path + ` tid=` + reTid + `\n`
		re := `^` + l + `tag=REQ_BEG ` + r + l + `tag=REQ_END code=` + test.code + ` dur=[0-9]+ ` + r + `$`
		if !regexp.MustCompile(re).Match(buf.Bytes()) {
			t.Fatalf("request log should match %q is %q", re, buf.Bytes())
		}
	}
}

func TestRelay6(t *testing.T) {
	var buf bytes.Buffer
	var l *Logger = New(NewTextHandler(&buf, NewOptions(LevelInfo, false, false)))

	mux := httpd.NewMux()
	mux.HandleRelay(l.Relay)
	mux.Handle("/200", http.MethodGet, func(s *httpd.Store) { s.W.WriteHeader(200) })
	mux.Handle("/400", http.MethodPost, func(s *httpd.Store) { s.W.WriteHeader(400) })
	mux.Handle("/403", http.MethodPut, func(s *httpd.Store) { s.W.WriteHeader(403) })
	mux.Handle("/404", http.MethodDelete, func(s *httpd.Store) { s.W.WriteHeader(404) })

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
	}
	for _, test := range tests {
		buf.Reset()
		requestDiscard(t, test.method, "http://"+server.Addr+test.path)

		reL := `time=` + reTextTime + ` level=INFO msg="" `
		reR := `ip=::1 method=` + test.method + ` path=` + test.path + ` tid=` + reTid + `\n`
		re := `^` + reL + `tag=REQ_BEG ` + reR + reL + `tag=REQ_END code=` + test.code + ` dur=[0-9]+ ` + reR + `$`
		if !regexp.MustCompile(re).Match(buf.Bytes()) {
			t.Fatalf("request log should match %q is %q", re, buf.Bytes())
		}
	}
}

func TestRelayRecover(t *testing.T) {
	var buf bytes.Buffer
	var l *Logger = New(NewTextHandler(&buf, NewOptions(LevelInfo, false, false)))

	mux := httpd.NewMux()
	mux.HandleRelay(l.Relay)
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
	reR := `ip=127.0.0.1 method=` + http.MethodGet + ` path=/panic tid=` + reTid + `\n`
	re := `^` + reL + `tag=REQ_BEG ` + reR +
		`time=` + reTextTime + ` level=ERROR msg="goroutine[^"]+" panic=expected tid=` + reTid + `\n` +
		reL + `tag=REQ_END code=500 dur=[0-9]+ ` + reR + `$`
	if !regexp.MustCompile(re).Match(buf.Bytes()) {
		t.Fatalf("request log should match %q is %q", re, buf.Bytes())
	}
}
