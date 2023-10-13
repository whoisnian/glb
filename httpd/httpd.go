// Package httpd implements a simple HTTP router with path parameters support.
package httpd

import (
	"crypto/rand"
	"encoding/base32"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
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
	storeID   uint64

	relayHandler  HandlerFunc
	routeNotFound *RouteInfo
}

// NewMux allocates and returns a new Mux.
func NewMux() *Mux {
	buf, prefix := make([]byte, 5), make([]byte, 9)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	base32.StdEncoding.Encode(prefix, buf)
	prefix[8] = '-'

	mux := &Mux{root: new(treeNode)}
	mux.HandleRelay(func(store *Store) { store.I.HandlerFunc(store) })
	mux.HandleNoRoute(func(store *Store) { store.Error404("404 not found") })
	mux.storePool.New = mux.newStoreWith(prefix)
	return mux
}

func (mux *Mux) newStoreWith(prefix []byte) func() any {
	return func() any {
		params := Params{V: make([]string, 0, mux.maxParams)}
		buf := make([]byte, 9, 32)
		copy(buf, prefix)
		return &Store{W: &ResponseWriter{}, P: &params, id: buf}
	}
}

// ServeHTTP dispatches the request to the matched handler.
func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	store := mux.storePool.Get().(*Store)
	store.W.Origin = w
	store.R = r
	store.id = strconv.AppendUint(store.id, atomic.AddUint64(&mux.storeID, 1), 36)

	if info := findRoute(mux.root, r.URL.Path, r.Method, store.P); info != nil {
		store.I = info
	} else {
		store.I = mux.routeNotFound
	}
	mux.relayHandler(store)

	store.W.Origin = nil
	store.W.Status = 0
	store.R = nil
	store.I = nil
	store.P.V = store.P.V[:0]
	store.id = store.id[:9]
	mux.storePool.Put(store)
}

// Handle registers the handler for the given routePath and method.
func (mux *Mux) Handle(path string, method string, handler HandlerFunc) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	info := newRouteInfo(path, method, handler)
	paramsCnt, err := parseRoute(mux.root, path, method, info)
	if err != nil {
		panic(err)
	}
	mux.maxParams = max(mux.maxParams, paramsCnt)
}

func (mux *Mux) HandleRelay(handler HandlerFunc) {
	mux.relayHandler = handler
}

func (mux *Mux) HandleNoRoute(handler HandlerFunc) {
	mux.routeNotFound = newRouteInfo("", "", handler)
}
