package httpd

import (
	"net/http"
	"net/url"
	"testing"
)

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
		{"/c", http.MethodGet, "/c"},
		{"/cc", http.MethodGet, "/cc"},
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

	root := new(treeNode)
	var maxParams int = 0
	for _, tt := range tests {
		info := newRouteInfo(tt.path, tt.method, func(*Store) {}, &[]HandlerFunc{})
		paramsCnt, err := parseRoute(root, tt.path, tt.method, info)
		if err != nil {
			t.Fatalf("parseRoute: %v", err)
		}
		maxParams = max(maxParams, paramsCnt)
	}

	params := Params{V: make([]string, 0, maxParams)}
	for _, tt := range tests {
		u, err := url.ParseRequestURI(tt.url)
		if err != nil {
			t.Fatalf("ParseRequestURI: %v", err)
		}

		method := tt.method
		if method == "*" {
			method = http.MethodConnect
		}
		params.V = params.V[:0]
		info := findRoute(root, u.Path, method, &params)
		if info == nil {
			t.Fatalf("routeInfo for %q not found", tt.url)
		}
		if info.Method != tt.method || info.Path != tt.path {
			t.Fatalf("url %q match %q %q, want %q %q", tt.url, info.Method, info.Path, tt.method, tt.path)
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

	root := new(treeNode)
	var maxParams int = 0
	for _, tt := range tests {
		info := newRouteInfo(tt.path, tt.method, func(*Store) {}, &[]HandlerFunc{})
		paramsCnt, err := parseRoute(root, tt.path, tt.method, info)
		if err != nil {
			t.Fatalf("parseRoute: %v", err)
		}
		maxParams = max(maxParams, paramsCnt)
	}

	params := Params{V: make([]string, 0, maxParams)}
	for _, tt := range tests {
		u, err := url.ParseRequestURI(tt.url)
		if err != nil {
			t.Fatalf("ParseRequestURI: %v", err)
		}

		params.V = params.V[:0]
		info := findRoute(root, u.Path, tt.method, &params)
		if info == nil {
			t.Fatalf("routeInfo for %q not found", tt.url)
		}
		store := &Store{P: &params, I: info}

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
