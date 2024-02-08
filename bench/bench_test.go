package bench

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
)

func init() {
	println("#GithubAPI Routes:", len(githubAPI))
	calcMem("HttpRouter", func() { githubHttpRouter = httpRouterLoad(githubAPI, false) })
	calcMem("Gin", func() { githubGin = ginLoad(githubAPI, false) })
	calcMem("Httpd", func() { githubHttpd = httpdLoad(githubAPI, false) })
	calcMem("Stdhttp", func() { githubStdhttp = stdhttpLoad(githubAPI, false) })
	println()
}

func calcMem(name string, load func()) {
	m := new(runtime.MemStats)

	runtime.GC()
	runtime.GC()
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(m)
	before := m.HeapAlloc

	load()

	runtime.GC()
	runtime.GC()
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(m)
	after := m.HeapAlloc
	println("   "+name+":", after-before, "Bytes")
}

var routers = []struct {
	name string
	load func(routes []route, test bool) http.Handler
}{
	{"HttpRouter", httpRouterLoad},
	{"Gin", ginLoad},
	{"Httpd", httpdLoad},
	{"Stdhttp", stdhttpLoad},
}

func TestRouters(t *testing.T) {
	for _, router := range routers {
		req, _ := http.NewRequest("GET", "/", nil)
		u := req.URL
		rq := u.RawQuery

		mux := router.load(githubAPI, true)
		for _, route := range githubAPI {
			w := httptest.NewRecorder()
			req.Method = route.method
			req.RequestURI = route.path
			u.Path = route.path
			u.RawQuery = rq
			mux.ServeHTTP(w, req)
			if w.Code != 200 || w.Body.String() != route.path {
				t.Errorf(
					"%s in API %s: %d - %s; expected %s %s\n",
					router.name, "GitHub", w.Code, w.Body.String(), route.method, route.path,
				)
			}
		}

	}
}

var (
	githubHttpd      http.Handler
	githubGin        http.Handler
	githubHttpRouter http.Handler
	githubStdhttp    http.Handler
)

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() http.Header         { return http.Header{} }
func (m *mockResponseWriter) Write(b []byte) (int, error) { return len(b), nil }
func (m *mockResponseWriter) WriteHeader(int)             {}

func benchRoutes(b *testing.B, router http.Handler, routes []route) {
	w := new(mockResponseWriter)
	r, _ := http.NewRequest("GET", "/", nil)
	u := r.URL
	rq := u.RawQuery

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			r.Method = route.method
			r.RequestURI = route.path
			u.Path = route.path
			u.RawQuery = rq
			router.ServeHTTP(w, r)
		}
	}
}

func BenchmarkHttpRouter_GithubAll(b *testing.B) {
	benchRoutes(b, githubHttpRouter, githubAPI)
}
func BenchmarkGinRouter_GithubAll(b *testing.B) {
	benchRoutes(b, githubGin, githubAPI)
}
func BenchmarkHttpd_GithubAll(b *testing.B) {
	benchRoutes(b, githubHttpd, githubAPI)
}
func BenchmarkStdhttp_GithubAll(b *testing.B) {
	benchRoutes(b, githubStdhttp, githubAPI)
}
