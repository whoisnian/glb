package logger_test

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"regexp"
	"testing"

	"github.com/whoisnian/glb/logger"
)

const (
	Rip4    = `[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`
	Rip6    = `([0-9a-fA-F]{0,4}:){1,7}[0-9a-fA-F]{0,4}`
	Rip     = `((` + Rip4 + `)|(` + Rip6 + `))`
	Rstatus = `\[[0-9]{3}\]`
	Rmethod = `(GET|HEAD|POST|PUT|DELETE|CONNECT|OPTIONS|TRACE|PATCH)`
	Rpath   = `/\S*`
	Rua     = "Go-http-client/1.1"
	Rcost   = `[0-9]+`
)

var (
	httpResp     = []byte("hello")
	reForRequest = regexp.MustCompile("^" + Rdate + " " + Rtime + " " + RplainLabel + " " + Rip + " " + Rstatus + " " + Rmethod + " " + Rpath + " " + Rua + " " + Rcost + `\n$`)
	reForPanic   = regexp.MustCompile("^" + Rdate + " " + Rtime + " " + RplainLabel + " panic: " + string(httpResp) + "\ngoroutine ")
)

func requestAndCheck(t *testing.T, url string, code int, body []byte) {
	if resp, err := http.Get(url); err != nil {
		t.Fatalf("request %v get err %v", url, err)
	} else if resp.StatusCode != code {
		t.Fatalf("request %v get status %d, want %d", url, resp.StatusCode, code)
	} else if data, _ := io.ReadAll(resp.Body); !bytes.Equal(data, body) {
		defer resp.Body.Close()
		t.Fatalf("request %v get body %q, want %q", url, data, body)
	}
}

func TestReq4(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	t.Cleanup(resetLogger)

	mux := http.NewServeMux()
	mux.HandleFunc("/200", func(w http.ResponseWriter, _ *http.Request) { w.Write(httpResp) })
	mux.HandleFunc("/400", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadRequest); w.Write(httpResp) })
	mux.HandleFunc("/500", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(httpResp)
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	server := &http.Server{Addr: ln.Addr().String(), Handler: logger.Req(mux)}
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	defer server.Shutdown(context.Background())

	buf.Reset()
	requestAndCheck(t, "http://"+server.Addr+"/200", http.StatusOK, httpResp)
	if !reForRequest.Match(buf.Bytes()) {
		t.Fatalf("request log should match %q is %q", reForRequest, buf.Bytes())
	}

	buf.Reset()
	requestAndCheck(t, "http://"+server.Addr+"/400", http.StatusBadRequest, httpResp)
	if !reForRequest.Match(buf.Bytes()) {
		t.Fatalf("request log should match %q is %q", reForRequest, buf.Bytes())
	}

	buf.Reset()
	requestAndCheck(t, "http://"+server.Addr+"/500", http.StatusInternalServerError, httpResp)
	if !reForRequest.Match(buf.Bytes()) {
		t.Fatalf("request log should match %q is %q", reForRequest, buf.Bytes())
	}
}

func TestReq6(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	t.Cleanup(resetLogger)

	mux := http.NewServeMux()
	mux.HandleFunc("/200", func(w http.ResponseWriter, _ *http.Request) { w.Write(httpResp) })
	mux.HandleFunc("/400", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadRequest); w.Write(httpResp) })
	mux.HandleFunc("/500", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(httpResp)
	})
	ln, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	server := &http.Server{Addr: ln.Addr().String(), Handler: logger.Req(mux)}
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	defer server.Shutdown(context.Background())

	buf.Reset()
	requestAndCheck(t, "http://"+server.Addr+"/200", http.StatusOK, httpResp)
	if !reForRequest.Match(buf.Bytes()) {
		t.Fatalf("request log should match %q is %q", reForRequest, buf.Bytes())
	}

	buf.Reset()
	requestAndCheck(t, "http://"+server.Addr+"/400", http.StatusBadRequest, httpResp)
	if !reForRequest.Match(buf.Bytes()) {
		t.Fatalf("request log should match %q is %q", reForRequest, buf.Bytes())
	}

	buf.Reset()
	requestAndCheck(t, "http://"+server.Addr+"/500", http.StatusInternalServerError, httpResp)
	if !reForRequest.Match(buf.Bytes()) {
		t.Fatalf("request log should match %q is %q", reForRequest, buf.Bytes())
	}
}

func TestRecovery(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	t.Cleanup(resetLogger)

	mux := http.NewServeMux()
	mux.HandleFunc("/panic", func(w http.ResponseWriter, _ *http.Request) { panic(string(httpResp)) })
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	server := &http.Server{Addr: ln.Addr().String(), Handler: logger.Recovery(mux)}
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	defer server.Shutdown(context.Background())

	buf.Reset()
	requestAndCheck(t, "http://"+server.Addr+"/panic", http.StatusInternalServerError, []byte(http.StatusText(http.StatusInternalServerError)+"\n"))
	if !reForPanic.Match(buf.Bytes()) {
		t.Fatalf("request log should match %q is %s", reForPanic, buf.Bytes())
	}
}

func TestRecoveryWithReq(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	t.Cleanup(resetLogger)

	mux := http.NewServeMux()
	mux.HandleFunc("/panic", func(w http.ResponseWriter, _ *http.Request) { panic(string(httpResp)) })
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	server := &http.Server{Addr: ln.Addr().String(), Handler: logger.Req(logger.Recovery(mux))}
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	defer server.Shutdown(context.Background())

	buf.Reset()
	requestAndCheck(t, "http://"+server.Addr+"/panic", http.StatusInternalServerError, []byte(http.StatusText(http.StatusInternalServerError)+"\n"))
	end := bytes.LastIndexByte(buf.Bytes()[:buf.Len()-1], '\n')
	if end < 0 {
		t.Fatalf("\\n not found in request log")
	}
	if !reForPanic.Match(buf.Bytes()[0 : end+1]) {
		t.Fatalf("request log should match %q is %s", reForPanic, buf.Bytes()[0:end+1])
	}
	if !reForRequest.Match(buf.Bytes()[end+1:]) {
		t.Fatalf("request log should match %q is %q", reForRequest, buf.Bytes()[end+1:])
	}
}
