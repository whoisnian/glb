package bench

import (
	"net/http"
	"regexp"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
	"github.com/whoisnian/glb/httpd"
)

var stdPatternReg = regexp.MustCompile(`/:([^/]+)`)

func init() {
	runtime.GOMAXPROCS(1)
	gin.SetMode(gin.ReleaseMode)
}

type route struct {
	method string
	path   string
}

func (r route) stdPattern() string {
	return r.method + " " + stdPatternReg.ReplaceAllString(r.path, `/{$1}`)
}

// httpd
func httpdHandler(_ *httpd.Store)     {}
func httpdHandlerTest(s *httpd.Store) { s.W.Write([]byte(s.R.RequestURI)) }
func httpdLoad(routes []route, test bool) http.Handler {
	h := httpdHandler
	if test {
		h = httpdHandlerTest
	}
	router := httpd.NewMux()
	for _, route := range routes {
		router.Handle(route.path, route.method, h)
	}
	return router
}

// Gin
func ginHandle(_ *gin.Context)     {}
func ginHandleTest(c *gin.Context) { c.Writer.Write([]byte(c.Request.RequestURI)) }
func ginLoad(routes []route, test bool) http.Handler {
	h := ginHandle
	if test {
		h = ginHandleTest
	}
	router := gin.New()
	for _, route := range routes {
		router.Handle(route.method, route.path, h)
	}
	return router
}

// HttpRouter
func httpRouterHandler(_ http.ResponseWriter, _ *http.Request, _ httprouter.Params) {}
func httpRouterHandlerTest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte(r.RequestURI))
}
func httpRouterLoad(routes []route, test bool) http.Handler {
	h := httpRouterHandler
	if test {
		h = httpRouterHandlerTest
	}
	router := httprouter.New()
	for _, route := range routes {
		router.Handle(route.method, route.path, h)
	}
	return router
}

// Stdhttp
func stdhttpHandler(_ http.ResponseWriter, _ *http.Request)     {}
func stdhttpHandlerTest(w http.ResponseWriter, r *http.Request) { w.Write([]byte(r.RequestURI)) }
func stdhttpLoad(routes []route, test bool) http.Handler {
	h := stdhttpHandler
	if test {
		h = stdhttpHandlerTest
	}
	router := http.NewServeMux()
	for _, route := range routes {
		router.HandleFunc(route.stdPattern(), h)
	}
	return router
}
