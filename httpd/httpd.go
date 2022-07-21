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
	http.MethodGet:     "/get",
	http.MethodHead:    "/head",
	http.MethodPost:    "/post",
	http.MethodPut:     "/put",
	http.MethodPatch:   "/patch",
	http.MethodDelete:  "/delete",
	http.MethodConnect: "/connect",
	http.MethodOptions: "/options",
	http.MethodTrace:   "/trace",
	MethodAll:          methodTagAll,
}

type nodeData struct {
	handler       HandlerFunc
	paramNameList []string
}

type routeNode struct {
	next map[string]*routeNode
	data *nodeData
}

func (node *routeNode) nextNodeOrNew(name string) (res *routeNode) {
	if res, ok := node.next[name]; ok {
		return res
	}
	if node.next == nil {
		node.next = make(map[string]*routeNode)
	}
	res = new(routeNode)
	node.next[name] = res
	return res
}

func (node *routeNode) methodNodeOrNil(method string) (res *routeNode) {
	if res, ok := node.next[methodTagMap[method]]; ok {
		return res
	}
	if res, ok := node.next[methodTagAll]; ok {
		return res
	}
	return nil
}

func parseRoute(node *routeNode, path string, method string) (*routeNode, []string, error) {
	methodTag, ok := methodTagMap[method]
	if !ok {
		return nil, nil, errors.New("invalid method " + method + " for routePath: " + path)
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
				return nil, nil, errors.New("invalid fragment :" + paramName + " in routePath: " + path)
			}
			paramNameList = append(paramNameList, paramName)
			node = node.nextNodeOrNew(routeParam)
		} else {
			node = node.nextNodeOrNew(path[left+1 : right])
		}
		left = right
	}

	if _, ok = node.next[methodTag]; ok {
		return nil, nil, errors.New("duplicate method " + method + " for routePath: " + path)
	}
	return node.nextNodeOrNew(methodTag), paramNameList, nil
}

// about trailing slash:
//   `/foo/bar`  will be matched by `/foo/bar`
//   `/foo/bar/` will be matched by `/foo/bar/:param` or `/foo/bar/*`
func findRoute(node *routeNode, path string, method string) (*routeNode, []string) {
	var paramValueList []string
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
			paramValueList = append(paramValueList, path[left+1:right])
			node = res
		} else if res, ok := node.next[routeParamAny]; ok {
			paramValueList = append(paramValueList, path[left+1:])
			node = res
			break
		} else {
			return nil, nil
		}
		left = right
	}
	return node.methodNodeOrNil(method), paramValueList
}

type Mux struct {
	mu   sync.RWMutex
	root *routeNode
}

// NewMux allocates and returns a new Mux.
func NewMux() *Mux {
	return &Mux{root: new(routeNode)}
}

// ServeHTTP dispatches the request to the matched handler.
func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	store := &Store{w, r, make(map[string]string)}

	node, paramValueList := findRoute(mux.root, r.URL.Path, r.Method)
	if node == nil {
		store.Error404("Route not found")
		return
	}

	for i := range node.data.paramNameList {
		store.m[node.data.paramNameList[i]] = paramValueList[i]
	}

	node.data.handler(store)
}

// Handle registers the handler for the given routePath and method.
func (mux *Mux) Handle(path string, method string, handler HandlerFunc) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	node, paramNameList, err := parseRoute(mux.root, path, method)
	if err != nil {
		panic(err)
	}
	node.data = &nodeData{handler, paramNameList}
}
