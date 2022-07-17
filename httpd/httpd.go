// Package httpd implements a simple HTTP router with path parameters support.
package httpd

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/whoisnian/glb/util/strutil"
)

const routeParam string = "/:param"
const routeParamAny string = "/:any"

var methodList = map[string]string{
	"GET":     "/get",
	"HEAD":    "/head",
	"POST":    "/post",
	"PUT":     "/put",
	"DELETE":  "/delete",
	"CONNECT": "/connect",
	"OPTIONS": "/options",
	"TRACE":   "/trace",
	"PATCH":   "/patch",
	"*":       "/*",
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
	methodTag := methodList[method]
	if res, ok := node.next[methodTag]; ok {
		return res
	}
	if res, ok := node.next[methodList["*"]]; ok {
		return res
	}
	return nil
}

func parseRoute(node *routeNode, path string, method string) (*routeNode, []string, error) {
	methodTag, ok := methodList[method]
	if !ok {
		return nil, nil, errors.New("invalid method " + method + " for routePath: " + path)
	}

	var paramNameList []string
	fragments := strings.Split(path, "/")
	for _, fragment := range fragments {
		if len(fragment) < 1 {
			continue
		} else if fragment == "*" {
			paramNameList = append(paramNameList, routeParamAny)
			node = node.nextNodeOrNew(routeParamAny)
			break
		} else if fragment[0] == ':' {
			paramName := fragment[1:]
			if paramName == "" || strutil.SliceContain(paramNameList, paramName) {
				return nil, nil, errors.New("invalid fragment " + fragment + " in routePath: " + path)
			}
			paramNameList = append(paramNameList, paramName)
			node = node.nextNodeOrNew(routeParam)
		} else {
			node = node.nextNodeOrNew(fragment)
		}
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
	fragments := strings.Split(path, "/")
	for i, fragment := range fragments {
		if len(fragment) < 1 && i < len(fragments)-1 { // check routeParam if current is last fragment
			continue
		} else if res, ok := node.next[fragment]; ok {
			node = res
		} else if res, ok := node.next[routeParam]; ok {
			paramValueList = append(paramValueList, fragment)
			node = res
		} else if res, ok := node.next[routeParamAny]; ok {
			paramValueList = append(paramValueList, strings.Join(fragments[i:], "/"))
			node = res
			break
		} else {
			return nil, nil
		}
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
