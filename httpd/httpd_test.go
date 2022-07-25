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

func TestRoute(t *testing.T) {
	tests := []struct {
		path   string
		method string
		url    string
	}{
		{"/", http.MethodGet, "/"},
		{"/*", http.MethodGet, "/a"},
		{"/aaa", http.MethodGet, "/aaa"},
		{"/aaa/bbb", http.MethodGet, "/aaa///bbb"},
		{"/ccc", http.MethodGet, "/ccc"},
		{"/ccc", http.MethodPost, "/ccc"},
		{"/ccc", http.MethodDelete, "/ccc"},
		{"/ddd", MethodAll, "/ddd"},
		{"/ddd", http.MethodPut, "/ddd"},
		{"/eee/:id", http.MethodPost, "/eee/10"},
		{"/fff/*", http.MethodPost, "/fff/any"},
		{"/ggg", http.MethodGet, "/ggg"},
		{"/ggg/*", http.MethodGet, "/ggg/"},
	}

	root := new(routeNode)
	var maxParams int = 0
	for _, tt := range tests {
		paramsCnt, err := parseRoute(root, tt.path, tt.method, createTestHandlerFunc(tt.method+tt.path))
		if err != nil {
			t.Fatalf("parseRoute: %v", err)
		}
		if paramsCnt > maxParams {
			maxParams = paramsCnt
		}
	}

	params := Params{V: make([]string, 0, maxParams)}
	for _, tt := range tests {
		u, err := url.ParseRequestURI(tt.url)
		if err != nil {
			t.Fatalf("ParseRequestURI: %v", err)
		}

		method := tt.method
		if method == "*" {
			method = "CONNECT"
		}
		params.V = params.V[:0]
		handler := findRoute(root, u.Path, method, &params)
		if handler == nil {
			t.Fatalf("routeNode for %q not found", tt.url)
		}
		w := &fakeResponseWriter{}
		store := &Store{w, &http.Request{}, &params}
		handler(store)
		if w.code != http.StatusOK || w.buf.String() != tt.method+tt.path {
			t.Fatalf("url %q match %q, want %q", tt.url, w.buf.String(), tt.method+tt.path)
		}
	}
}

func TestRouteParam(t *testing.T) {
	tests := []struct {
		path   string
		method string
		url    string
		paramK []string
		paramV []string
	}{
		{"/*", http.MethodGet, "/", []string{routeParamAny}, []string{""}},
		{"/aaa/:id", http.MethodGet, "/aaa/10", []string{"id"}, []string{"10"}},
		{"/bbb/:id", http.MethodGet, "/bbb/10", []string{"id"}, []string{"10"}},
		{"/bbb/ccc", http.MethodGet, "/bbb/ccc", []string{"none"}, []string{""}},
		{"/ccc/:id1/:id2/ddd", http.MethodGet, "/ccc/10/20/ddd", []string{"id1", "id2"}, []string{"10", "20"}},
		{"/ccc/:id1/ddd/:id2", http.MethodGet, "/ccc/10/ddd/20", []string{"id1", "id2"}, []string{"10", "20"}},
		{"/eee/:id", http.MethodGet, "/eee/10", []string{"id"}, []string{"10"}},
		{"/eee/:id1/:id2", http.MethodGet, "/eee/10/20", []string{"id1", "id2"}, []string{"10", "20"}},
		{"/fff/*", http.MethodGet, "/fff/1/2/3", []string{routeParamAny}, []string{"1/2/3"}},
	}

	root := new(routeNode)
	var maxParams int = 0
	for _, tt := range tests {
		paramsCnt, err := parseRoute(root, tt.path, tt.method, createTestHandlerFunc(tt.method+tt.path))
		if err != nil {
			t.Fatalf("parseRoute: %v", err)
		}
		if paramsCnt > maxParams {
			maxParams = paramsCnt
		}
	}

	params := Params{V: make([]string, 0, maxParams)}
	for _, tt := range tests {
		u, err := url.ParseRequestURI(tt.url)
		if err != nil {
			t.Fatalf("ParseRequestURI: %v", err)
		}

		params.V = params.V[:0]
		node := findRoute(root, u.Path, tt.method, &params)
		if node == nil {
			t.Fatalf("routeNode for %q not found", tt.url)
		}
		store := &Store{P: &params}

		var v string
		for i, k := range tt.paramK {
			if k == routeParamAny {
				v = store.RouteParamAny()
			} else {
				v = store.RouteParam(k)
			}
			if v != tt.paramV[i] {
				t.Fatalf("RouteParam(%q) = %q, want %q", k, v, tt.paramV[i])
			}
		}
	}
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
		{"/", http.MethodGet, 404, "404 page not found\n"},
		{"/aaa", http.MethodGet, 200, "get_aaa"},
		{"/aaa", http.MethodPost, 404, "404 page not found\n"},
		{"/aaa/", http.MethodGet, 404, "404 page not found\n"},
		{"/bbb", http.MethodPost, 404, "404 page not found\n"},
		{"/bbb/", http.MethodPost, 200, "post_bbb"},
		{"/bbb/10", http.MethodPost, 200, "post_bbb"},
		{"/ccc", http.MethodGet, 200, "get_ccc"},
		{"/ccc", http.MethodConnect, 200, "any_ccc"},
		{"/ddd", http.MethodPut, 404, "404 page not found\n"},
		{"/ddd/", http.MethodPut, 200, "put_ddd"},
		{"/ddd/10", http.MethodPut, 200, "put_ddd"},
		{"/ddd/eee", http.MethodPut, 200, "put_ddd_eee"},
		{"/fff", http.MethodGet, 404, "404 page not found\n"},
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
