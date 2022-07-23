// Package httpd implements a simple HTTP router with path parameters support.
package httpd

import (
	"errors"
	"net/http"
	"sync"

	"github.com/whoisnian/glb/util/strutil"
)

const (
	MethodGet     = http.MethodGet
	MethodHead    = http.MethodHead
	MethodPost    = http.MethodPost
	MethodPut     = http.MethodPut
	MethodPatch   = http.MethodPatch
	MethodDelete  = http.MethodDelete
	MethodConnect = http.MethodConnect
	MethodOptions = http.MethodOptions
	MethodTrace   = http.MethodTrace
	MethodAll     = "*"
)

const routeParam string = "/:param"
const routeParamAny string = "/:any"
const methodTagAll string = "/*"

var methodTagMap = map[string]string{
	MethodGet:     "/get",
	MethodHead:    "/head",
	MethodPost:    "/post",
	MethodPut:     "/put",
	MethodPatch:   "/patch",
	MethodDelete:  "/delete",
	MethodConnect: "/connect",
	MethodOptions: "/options",
	MethodTrace:   "/trace",
	MethodAll:     methodTagAll,
}

type routeNode struct {
	next          map[string]*routeNode
	handler       HandlerFunc
	paramNameList []string
}

func (node *routeNode) nextNodeOrNew(name string) (resNode *routeNode) {
	if resNode, ok := node.next[name]; ok {
		return resNode
	}
	if node.next == nil {
		node.next = make(map[string]*routeNode)
	}
	resNode = new(routeNode)
	node.next[name] = resNode
	return resNode
}

func (node *routeNode) methodNodeOrNil(method string) (resNode *routeNode) {
	if resNode, ok := node.next[methodTagMap[method]]; ok {
		return resNode
	}
	if resNode, ok := node.next[methodTagAll]; ok {
		return resNode
	}
	return nil
}

func parseRoute(node *routeNode, path string, method string, handler HandlerFunc) (paramsCnt int, err error) {
	methodTag, ok := methodTagMap[method]
	if !ok {
		return 0, errors.New("invalid method " + method + " for routePath: " + path)
	}

	var paramNameList []string
	var length, left, right int = len(path), 0, 0
	for ; right <= length; right++ {
		if right < length && path[right] != '/' {
			continue
		}
		if right-left < 2 {
			// continue
		} else if path[left+1:right] == "*" {
			paramNameList = append(paramNameList, routeParamAny)
			node = node.nextNodeOrNew(routeParamAny)
			break
		} else if path[left+1] == ':' {
			paramName := path[left+2 : right]
			if paramName == "" || strutil.SliceContain(paramNameList, paramName) {
				return 0, errors.New("invalid fragment :" + paramName + " in routePath: " + path)
			}
			paramNameList = append(paramNameList, paramName)
			node = node.nextNodeOrNew(routeParam)
		} else {
			node = node.nextNodeOrNew(path[left+1 : right])
		}
		left = right
	}

	if _, ok = node.next[methodTag]; ok {
		return 0, errors.New("duplicate method " + method + " for routePath: " + path)
	}
	node = node.nextNodeOrNew(methodTag)
	node.handler = handler
	node.paramNameList = paramNameList
	return len(paramNameList), nil
}

// about trailing slash:
//   `/foo/bar`  will be matched by `/foo/bar`
//   `/foo/bar/` will be matched by `/foo/bar/:param` or `/foo/bar/*`
func findRoute(node *routeNode, path string, method string, params *Params) (handler HandlerFunc) {
	var length, left, right int = len(path), 0, 0
	for ; right <= length; right++ {
		if right < length && path[right] != '/' {
			continue
		}
		if right-left < 2 && right < length { // check routeParam if current is last fragment
			// continue
		} else if res, ok := node.next[path[left+1:right]]; ok {
			node = res
		} else if res, ok := node.next[routeParam]; ok {
			i := len(params.V)
			params.V = params.V[:i+1]
			params.V[i] = path[left+1 : right]
			node = res
		} else if res, ok := node.next[routeParamAny]; ok {
			i := len(params.V)
			params.V = params.V[:i+1]
			params.V[i] = path[left+1:]
			node = res
			break
		} else {
			return nil
		}
		left = right
	}
	if node = node.methodNodeOrNil(method); node != nil {
		params.K = node.paramNameList
		return node.handler
	} else {
		return nil
	}
}

type Mux struct {
	mu   sync.RWMutex
	root *routeNode

	maxParams int
	storePool sync.Pool

	NotFound http.HandlerFunc
}

// NewMux allocates and returns a new Mux.
func NewMux() *Mux {
	mux := &Mux{root: new(routeNode), NotFound: http.NotFound}
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

	handler := findRoute(mux.root, r.URL.Path, r.Method, store.P)
	if handler == nil {
		mux.NotFound(w, r)
	} else {
		handler(store)
	}

	mux.storePool.Put(store)
}

// Handle registers the handler for the given routePath and method.
func (mux *Mux) Handle(path string, method string, handler HandlerFunc) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	paramsCnt, err := parseRoute(mux.root, path, method, handler)
	if err != nil {
		panic(err)
	}

	if paramsCnt > mux.maxParams {
		mux.maxParams = paramsCnt
	}
}
