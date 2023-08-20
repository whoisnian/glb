// Package httpd implements a simple HTTP router with path parameters support.
package httpd

import (
	"net/http"
	"sync"
)

const MethodAll string = "*"

var methodTagMap = map[string]string{
	http.MethodGet:     "/get",
	http.MethodHead:    "/head",
	http.MethodPost:    "/post",
	http.MethodPut:     "/put",
	http.MethodPatch:   "/patch",
	http.MethodDelete:  "/delete",
	http.MethodConnect: "/connect",
	http.MethodOptions: "/options",
	http.MethodTrace:   "/trace",
	MethodAll:          "/*",
}

type Mux struct {
	mu   sync.RWMutex
	root *treeNode

	maxParams int
	storePool sync.Pool

	notFoundRoute *RouteInfo
}

// NewMux allocates and returns a new Mux.
func NewMux() *Mux {
	mux := &Mux{root: new(treeNode)}
	mux.HandleNotFound(NotFoundHandler)
	mux.storePool.New = mux.newStore
	return mux
}

func (mux *Mux) newStore() any {
	params := Params{V: make([]string, 0, mux.maxParams)}
	return &Store{P: &params}
}

// ServeHTTP dispatches the request to the matched handler.
func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	store := mux.storePool.Get().(*Store)
	store.W = w
	store.R = r
	store.P.V = store.P.V[:0]

	info := findRoute(mux.root, r.URL.Path, r.Method, store.P)
	if info == nil {
		info = mux.notFoundRoute
	}
	store.I = info
	info.HandlerFunc(store)

	mux.storePool.Put(store)
}

// Handle registers the handler for the given routePath and method.
func (mux *Mux) Handle(path string, method string, handler HandlerFunc) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	info := newRouteInfo(method, path, handler)
	paramsCnt, err := parseRoute(mux.root, path, method, info)
	if err != nil {
		panic(err)
	}
	mux.maxParams = max(mux.maxParams, paramsCnt)
}

func (mux *Mux) HandleNotFound(handler HandlerFunc) {
	mux.notFoundRoute = newRouteInfo("", "", handler)
}

func NotFoundHandler(store *Store) {
	store.Error404("404 not found")
}
