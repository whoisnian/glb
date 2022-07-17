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
		{"/aaa", "GET", "/aaa"},
		{"/aaa/bbb", "GET", "/aaa///bbb"},
		{"/ccc", "GET", "/ccc"},
		{"/ccc", "POST", "/ccc"},
		{"/ccc", "DELETE", "/ccc"},
		{"/ddd", "*", "/ddd"},
		{"/ddd", "PUT", "/ddd"},
		{"/eee/:id", "POST", "/eee/10"},
		{"/fff/*", "POST", "/fff/any"},
		{"/ggg", "GET", "/ggg"},
		{"/ggg/*", "GET", "/ggg/"},
	}

	root := new(routeNode)
	for _, tt := range tests {
		node, paramNameList, err := parseRoute(root, tt.path, tt.method)
		if err != nil {
			t.Fatalf("parseRoute: %v", err)
		}
		node.data = &nodeData{createTestHandlerFunc(tt.method + tt.path), paramNameList}
	}

	for _, tt := range tests {
		u, err := url.ParseRequestURI(tt.url)
		if err != nil {
			t.Fatalf("ParseRequestURI: %v", err)
		}

		method := tt.method
		if method == "*" {
			method = "CONNECT"
		}
		node, paramValueList := findRoute(root, u.Path, method)
		if node == nil {
			t.Fatalf("routeNode for %q not found", tt.url)
		}
		w := &fakeResponseWriter{}
		store := &Store{w, &http.Request{}, make(map[string]string)}
		for i := range node.data.paramNameList {
			store.m[node.data.paramNameList[i]] = paramValueList[i]
		}

		node.data.handler(store)
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
		{"/aaa/:id", "GET", "/aaa/10", []string{"id"}, []string{"10"}},
		{"/bbb/:id", "GET", "/bbb/10", []string{"id"}, []string{"10"}},
		{"/bbb/ccc", "GET", "/bbb/ccc", []string{"none"}, []string{""}},
		{"/ccc/:id1/:id2/ddd", "GET", "/ccc/10/20/ddd", []string{"id1", "id2"}, []string{"10", "20"}},
		{"/ccc/:id1/ddd/:id2", "GET", "/ccc/10/ddd/20", []string{"id1", "id2"}, []string{"10", "20"}},
		{"/eee/:id", "GET", "/eee/10", []string{"id"}, []string{"10"}},
		{"/eee/:id1/:id2", "GET", "/eee/10/20", []string{"id1", "id2"}, []string{"10", "20"}},
		{"/fff/*", "GET", "/fff/1/2/3", []string{routeParamAny}, []string{"1/2/3"}},
	}

	root := new(routeNode)
	for _, tt := range tests {
		node, paramNameList, err := parseRoute(root, tt.path, tt.method)
		if err != nil {
			t.Fatalf("parseRoute: %v", err)
		}
		node.data = &nodeData{createTestHandlerFunc(tt.method + tt.path), paramNameList}
	}

	for _, tt := range tests {
		u, err := url.ParseRequestURI(tt.url)
		if err != nil {
			t.Fatalf("ParseRequestURI: %v", err)
		}

		node, paramValueList := findRoute(root, u.Path, tt.method)
		if node == nil {
			t.Fatalf("routeNode for %q not found", tt.url)
		}
		store := &Store{nil, nil, make(map[string]string)}
		for i := range node.data.paramNameList {
			store.m[node.data.paramNameList[i]] = paramValueList[i]
		}

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
		{"duplicatedRoute", "/aaa", "GET"},
		{"duplicatedParam", "/bbb/:id/:id", "GET"},
		{"invalidPath", "/ccc/:/", "GET"},
		{"invalidMethod", "/ddd", "GETT"},
	}

	mux := NewMux()
	mux.Handle("/aaa", "GET", noop)
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
		{"/aaa", "GET", "get_aaa"},
		{"/bbb/:id", "POST", "post_bbb"},
		{"/ccc", "*", "any_ccc"},
		{"/ccc", "GET", "get_ccc"},
		{"/ddd/*", "PUT", "put_ddd"},
		{"/ddd/eee", "PUT", "put_ddd_eee"},
	}
	tests := []struct {
		url    string
		method string
		code   int
		mark   string
	}{
		{"/aaa", "GET", 200, "get_aaa"},
		{"/aaa", "POST", 404, "Route not found\n"},
		{"/aaa/", "GET", 404, "Route not found\n"},
		{"/bbb", "POST", 404, "Route not found\n"},
		{"/bbb/", "POST", 200, "post_bbb"},
		{"/bbb/10", "POST", 200, "post_bbb"},
		{"/ccc", "GET", 200, "get_ccc"},
		{"/ccc", "CONNECT", 200, "any_ccc"},
		{"/ddd", "PUT", 404, "Route not found\n"},
		{"/ddd/", "PUT", 200, "put_ddd"},
		{"/ddd/10", "PUT", 200, "put_ddd"},
		{"/ddd/eee", "PUT", 200, "put_ddd_eee"},
		{"/fff", "GET", 404, "Route not found\n"},
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
