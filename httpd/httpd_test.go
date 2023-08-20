package httpd_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/whoisnian/glb/httpd"
)

func TestHandlePanic(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		method string
	}{
		{"duplicatedRoute", "/aaa", http.MethodGet},
		{"duplicatedParam", "/bbb/:id/:id", http.MethodGet},
		{"invalidPath", "/ccc/:/", http.MethodGet},
		{"invalidMethod", "/ddd", "GETT"},
	}

	mux := httpd.NewMux()
	mux.Handle("/aaa", http.MethodGet, func(*httpd.Store) {})
	for _, tt := range tests {
		subtest := func(t *testing.T) {
			defer func() { _ = recover() }()
			mux.Handle(tt.path, tt.method, func(*httpd.Store) {})
			t.Fatalf("Handle(%q, %q) should panic", tt.method, tt.path)
		}
		if !t.Run(tt.name, subtest) {
			break
		}
	}
}

func TestHandleNotFound(t *testing.T) {
	tests := []struct {
		url    string
		method string
		code   int
	}{
		{"/", http.MethodGet, 404},
		{"/aaa", http.MethodPost, 404},
		{"/bbb/", http.MethodPut, 404},
		{"/ccc///ddd", http.MethodDelete, 404},
	}

	mark := "TestHandleNotFound"

	mux := httpd.NewMux()
	mux.HandleNotFound(func(s *httpd.Store) {
		s.W.WriteHeader(404)
		s.W.Write([]byte(mark))
	})
	for _, tt := range tests {
		u, err := url.ParseRequestURI(tt.url)
		if err != nil {
			t.Fatalf("ParseRequestURI: %v", err)
		}

		w := &fakeResponseWriter{header: make(http.Header)}
		mux.ServeHTTP(w, &http.Request{Method: tt.method, URL: u})
		if w.code != tt.code || w.buf.String() != mark {
			t.Fatalf("ServeHTTP(%q) return %d %q, want %d %q", tt.url, w.code, w.buf.String(), tt.code, mark)
		}
	}
}

func TestServeHTTP(t *testing.T) {
	routes := []struct {
		path   string
		method string
		mark   string
	}{
		{"/aaa", http.MethodGet, "get_aaa"},
		{"/bbb/:id", http.MethodPost, "post_bbb"},
		{"/ccc", httpd.MethodAll, "any_ccc"},
		{"/ccc", http.MethodGet, "get_ccc"},
		{"/ddd/*", http.MethodPut, "put_ddd"},
		{"/ddd/eee", http.MethodPut, "put_ddd_eee"},
	}
	tests := []struct {
		url    string
		method string
		code   int
		mark   string
	}{
		{"/", http.MethodGet, 404, "404 not found\n"},
		{"/aaa", http.MethodGet, 200, "get_aaa"},
		{"/aaa", http.MethodPost, 404, "404 not found\n"},
		{"/aaa/", http.MethodGet, 404, "404 not found\n"},
		{"/bbb", http.MethodPost, 404, "404 not found\n"},
		{"/bbb/", http.MethodPost, 200, "post_bbb"},
		{"/bbb/10", http.MethodPost, 200, "post_bbb"},
		{"/ccc", http.MethodGet, 200, "get_ccc"},
		{"/ccc", http.MethodConnect, 200, "any_ccc"},
		{"/ddd", http.MethodPut, 404, "404 not found\n"},
		{"/ddd/", http.MethodPut, 200, "put_ddd"},
		{"/ddd/10", http.MethodPut, 200, "put_ddd"},
		{"/ddd/eee", http.MethodPut, 200, "put_ddd_eee"},
		{"/fff", http.MethodGet, 404, "404 not found\n"},
	}

	mux := httpd.NewMux()
	for _, tt := range routes {
		tmp := []byte(tt.mark) // avoid sharing loop variable in anonymous function
		mux.Handle(tt.path, tt.method, func(s *httpd.Store) { s.W.Write(tmp) })
	}

	for _, tt := range tests {
		u, err := url.ParseRequestURI(tt.url)
		if err != nil {
			t.Fatalf("ParseRequestURI: %v", err)
		}

		w := &fakeResponseWriter{header: make(http.Header)}
		mux.ServeHTTP(w, &http.Request{Method: tt.method, URL: u})
		if w.code != tt.code || w.buf.String() != tt.mark {
			t.Fatalf("ServeHTTP(%q) return %d %q, want %d %q", tt.url, w.code, w.buf.String(), tt.code, tt.mark)
		}
	}
}
