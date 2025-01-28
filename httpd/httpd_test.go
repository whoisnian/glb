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

func TestHandleMiddleware(t *testing.T) {
	tests := []struct {
		path    string
		method  string
		content string
	}{
		{"/", http.MethodGet, "get /"},
		{"/aaa", http.MethodPost, "post /aaa"},
		{"/aa/bb", http.MethodPut, "put /aa/bb"},
		{"/aa/bb", http.MethodDelete, "delete /aa/bb"},
	}

	mark0 := "TestHandleMiddleware"
	mark1st := mark0 + "_1st_ "
	mark1ed := mark0 + "_1ed_ "
	mark2st := mark0 + "_2st_ "
	mark2ed := mark0 + "_2ed_ "

	mux := httpd.NewMux()
	mux.HandleMiddleware(func(s *httpd.Store) {
		s.W.Write([]byte(mark1st))
		s.Next()
		s.W.Write([]byte(mark1ed))
	})
	mux.HandleMiddleware(func(s *httpd.Store) {
		s.W.Write([]byte(mark2st))
		s.Next()
		s.W.Write([]byte(mark2ed))
	})
	for _, tt := range tests {
		data := []byte(tt.content)
		mux.Handle(tt.path, tt.method, func(s *httpd.Store) { s.W.Write(data) })
	}
	for _, tt := range tests {
		u, err := url.ParseRequestURI(tt.path)
		if err != nil {
			t.Fatalf("ParseRequestURI: %v", err)
		}

		w := &fakeResponseWriter{header: make(http.Header)}
		mux.ServeHTTP(w, &http.Request{Method: tt.method, URL: u})
		if want := mark1st + mark2st + tt.content + mark2ed + mark1ed; w.buf.String() != want {
			t.Fatalf("ServeHTTP(%q) return %q, want %q", tt.path, w.buf.String(), want)
		}
	}
}

func TestHandleNoRoute(t *testing.T) {
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

	mark := "TestHandleNoRoute"

	mux := httpd.NewMux()
	mux.HandleNoRoute(func(s *httpd.Store) {
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

func routeInfoStr(r httpd.RouteInfo) string { return r.Path + r.Method + r.HandlerName }
func testRouteInfo0(s *httpd.Store)         { s.W.Write([]byte(routeInfoStr(*s.I))) }
func testRouteInfo1(s *httpd.Store)         { s.W.Write([]byte(routeInfoStr(*s.I))) }
func testRouteInfo2(s *httpd.Store)         { s.W.Write([]byte(routeInfoStr(*s.I))) }
func testRouteInfo3(s *httpd.Store)         { s.W.Write([]byte(routeInfoStr(*s.I))) }
func testRouteInfo4(s *httpd.Store)         { s.W.Write([]byte(routeInfoStr(*s.I))) }
func testRouteInfo5(s *httpd.Store)         { s.W.Write([]byte(routeInfoStr(*s.I))) }

func TestRouteInfo(t *testing.T) {
	routes := []httpd.RouteInfo{
		{"", httpd.MethodAll, "github.com/whoisnian/glb/httpd_test.testRouteInfo0", testRouteInfo0, &[]httpd.HandlerFunc{}},
		{"/aaa", http.MethodGet, "github.com/whoisnian/glb/httpd_test.testRouteInfo1", testRouteInfo1, &[]httpd.HandlerFunc{}},
		{"/aaa", httpd.MethodAll, "github.com/whoisnian/glb/httpd_test.testRouteInfo2", testRouteInfo2, &[]httpd.HandlerFunc{}},
		{"/bbb/:id", http.MethodPost, "github.com/whoisnian/glb/httpd_test.testRouteInfo3", testRouteInfo3, &[]httpd.HandlerFunc{}},
		{"/ccc/*", http.MethodPut, "github.com/whoisnian/glb/httpd_test.testRouteInfo4", testRouteInfo4, &[]httpd.HandlerFunc{}},
		{"/ccc/ddd", http.MethodPut, "github.com/whoisnian/glb/httpd_test.testRouteInfo5", testRouteInfo5, &[]httpd.HandlerFunc{}},
	}
	tests := []struct {
		url    string
		method string
		want   string
	}{
		{"/", http.MethodGet, routeInfoStr(routes[0])},
		{"/aaa", http.MethodGet, routeInfoStr(routes[1])},
		{"/aaa", http.MethodPost, routeInfoStr(routes[2])},
		{"/bbb", http.MethodPost, routeInfoStr(routes[0])},
		{"/bbb/10", http.MethodPost, routeInfoStr(routes[3])},
		{"/ccc", http.MethodPut, routeInfoStr(routes[0])},
		{"/ccc/", http.MethodPut, routeInfoStr(routes[4])},
		{"/ccc/10", http.MethodPut, routeInfoStr(routes[4])},
		{"/ccc/ddd", http.MethodPut, routeInfoStr(routes[5])},
		{"/eee", http.MethodGet, routeInfoStr(routes[0])},
	}
	mux := httpd.NewMux()
	mux.HandleNoRoute(testRouteInfo0)
	for i := 1; i < len(routes); i++ {
		mux.Handle(routes[i].Path, routes[i].Method, routes[i].HandlerFunc)
	}

	for _, tt := range tests {
		u, err := url.ParseRequestURI(tt.url)
		if err != nil {
			t.Fatalf("ParseRequestURI: %v", err)
		}

		w := &fakeResponseWriter{header: make(http.Header)}
		mux.ServeHTTP(w, &http.Request{Method: tt.method, URL: u})
		if w.buf.String() != tt.want {
			t.Fatalf("RouteInfo return %s, want %s", w.buf.String(), tt.want)
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
