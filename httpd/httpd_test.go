package httpd

import (
	"net/http"
	"net/url"
	"testing"
)

func noop(*Store) {}
func createTestHandlerFunc(mark string) HandlerFunc {
	return func(s *Store) { s.W.Write([]byte(mark)) }
}

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

	mux := NewMux()
	mux.Handle("/aaa", http.MethodGet, noop)
	for _, tt := range tests {
		subtest := func(t *testing.T) {
			defer func() { _ = recover() }()
			mux.Handle(tt.path, tt.method, noop)
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
	resMark := mark + "\n"

	mux := NewMux()
	mux.HandleNotFound(func(store *Store) { store.Error404(mark) })
	for _, tt := range tests {
		u, err := url.ParseRequestURI(tt.url)
		if err != nil {
			t.Fatalf("ParseRequestURI: %v", err)
		}

		w := &fakeResponseWriter{header: make(http.Header)}
		mux.ServeHTTP(w, &http.Request{Method: tt.method, URL: u})
		if w.code != tt.code || w.buf.String() != resMark {
			t.Fatalf("ServeHTTP(%q) return %d %q, want %d %q", tt.url, w.code, w.buf.String(), tt.code, resMark)
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
		{"/ccc", MethodAll, "any_ccc"},
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

	mux := NewMux()
	for _, tt := range routes {
		mux.Handle(tt.path, tt.method, createTestHandlerFunc(tt.mark))
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
