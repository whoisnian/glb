package logger_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"testing"

	"github.com/whoisnian/glb/logger"
)

const (
	Rip     = `[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`
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
		t.Fatalf("request %v get body %q, want %q", url, data, body)
	}
}

func TestReq(t *testing.T) {
	var stdout bytes.Buffer
	logger.SetOutput(&stdout, nil)
	t.Cleanup(resetLogger)

	mux := http.NewServeMux()
	mux.HandleFunc("/200", func(w http.ResponseWriter, _ *http.Request) { w.Write(httpResp) })
	mux.HandleFunc("/400", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadRequest); w.Write(httpResp) })
	mux.HandleFunc("/500", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(httpResp)
	})
	server := &http.Server{Addr: ":8000", Handler: logger.Req(mux)}
	running := make(chan struct{})
	go func() {
		close(running)
		server.ListenAndServe()
	}()
	t.Cleanup(func() { server.Shutdown(context.Background()) })
	<-running

	stdout.Reset()
	requestAndCheck(t, "http://127.0.0.1:8000/200", http.StatusOK, httpResp)
	if !reForRequest.Match(stdout.Bytes()) {
		t.Fatalf("request log should match %q is %q", reForRequest, stdout.Bytes())
	}

	stdout.Reset()
	requestAndCheck(t, "http://127.0.0.1:8000/400", http.StatusBadRequest, httpResp)
	if !reForRequest.Match(stdout.Bytes()) {
		t.Fatalf("request log should match %q is %q", reForRequest, stdout.Bytes())
	}

	stdout.Reset()
	requestAndCheck(t, "http://127.0.0.1:8000/500", http.StatusInternalServerError, httpResp)
	if !reForRequest.Match(stdout.Bytes()) {
		t.Fatalf("request log should match %q is %q", reForRequest, stdout.Bytes())
	}
}

func TestRecovery(t *testing.T) {
	var stderr bytes.Buffer
	logger.SetOutput(nil, &stderr)
	t.Cleanup(resetLogger)

	mux := http.NewServeMux()
	mux.HandleFunc("/panic", func(w http.ResponseWriter, _ *http.Request) { panic(string(httpResp)) })
	server := &http.Server{Addr: ":8000", Handler: logger.Recovery(mux)}
	running := make(chan struct{})
	go func() {
		close(running)
		server.ListenAndServe()
	}()
	t.Cleanup(func() { server.Shutdown(context.Background()) })
	<-running

	stderr.Reset()
	requestAndCheck(t, "http://127.0.0.1:8000/panic", http.StatusInternalServerError, []byte(http.StatusText(http.StatusInternalServerError)+"\n"))
	if !reForPanic.Match(stderr.Bytes()) {
		t.Fatalf("request log should match %q is %s", reForPanic, stderr.Bytes())
	}
}

func TestRecoveryWithReq(t *testing.T) {
	var stdout, stderr bytes.Buffer
	logger.SetOutput(&stdout, &stderr)
	t.Cleanup(resetLogger)

	mux := http.NewServeMux()
	mux.HandleFunc("/panic", func(w http.ResponseWriter, _ *http.Request) { panic(string(httpResp)) })
	server := &http.Server{Addr: ":8000", Handler: logger.Req(logger.Recovery(mux))}
	running := make(chan struct{})
	go func() {
		close(running)
		server.ListenAndServe()
	}()
	t.Cleanup(func() { server.Shutdown(context.Background()) })
	<-running

	stdout.Reset()
	stderr.Reset()
	requestAndCheck(t, "http://127.0.0.1:8000/panic", http.StatusInternalServerError, []byte(http.StatusText(http.StatusInternalServerError)+"\n"))
	if !reForRequest.Match(stdout.Bytes()) {
		t.Fatalf("request log should match %q is %q", reForRequest, stdout.Bytes())
	}
	if !reForPanic.Match(stderr.Bytes()) {
		t.Fatalf("request log should match %q is %s", reForPanic, stderr.Bytes())
	}
}
