// Package httpd implements a simple HTTP router with path parameters support.
package httpd

import (
	"errors"
	"net/http"
	"net/url"
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

func parseRoute(node *routeNode, path string, method string) (*routeNode, []string) {
	var paramNameList []string
	fragments := strings.Split(path, "/")
	for i := range fragments {
		if len(fragments[i]) < 1 {
			continue
		} else if fragments[i] == "*" {
			paramNameList = append(paramNameList, routeParamAny)
			node = node.nextNodeOrNew(routeParamAny)
			break
		} else if fragments[i][0] == ':' {
			paramName := fragments[i][1:]
			if paramName == "" || strutil.SliceContain(paramNameList, paramName) {
				panic(errors.New("Invalid fragment '" + fragments[i] + "' in routePath: '" + path + "'"))
			}
			paramNameList = append(paramNameList, paramName)
			node = node.nextNodeOrNew(routeParam)
		} else {
			node = node.nextNodeOrNew(fragments[i])
		}
	}

	methodTag, ok := methodList[method]
	if !ok {
		panic(errors.New("Invalid method '" + method + "' for routePath: '" + path + "'"))
	}
	if _, ok = node.next[methodTag]; ok {
		panic(errors.New("Duplicate method '" + method + "' for routePath: '" + path + "'"))
	}
	return node.nextNodeOrNew(methodTag), paramNameList
}

// about trailing slash:
//   use `/foo/bar` to match `/foo/bar`
//   use `/foo/bar/*` to match `/foo/bar/`
func findRoute(node *routeNode, path string, method string) (*routeNode, []string) {
	var paramValueList []string
	fragments := strings.Split(path, "/")
	for i := range fragments {
		if len(fragments[i]) < 1 && i < len(fragments)-1 {
			continue
		} else if res, ok := node.next[fragments[i]]; ok {
			node = res
		} else if res, ok := node.next[routeParam]; ok {
			value, err := url.PathUnescape(fragments[i])
			if err != nil {
				// logger.Error("Invalid fragment '", fragments[i], "' in routePath: '", path, "'")
				return nil, nil
			}
			paramValueList = append(paramValueList, value)
			node = res
		} else if res, ok := node.next[routeParamAny]; ok {
			value, err := url.PathUnescape(strings.Join(fragments[i:], "/"))
			if err != nil {
				// logger.Error("Invalid fragment '", fragments[i], "' in routePath: '", path, "'")
				return nil, nil
			}
			paramValueList = append(paramValueList, value)
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

func NewMux() *Mux {
	return &Mux{root: new(routeNode)}
}

func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	store := &Store{w, r, make(map[string]string)}

	node, paramValueList := findRoute(mux.root, r.URL.EscapedPath(), r.Method)
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

	node, paramNameList := parseRoute(mux.root, path, method)
	node.data = &nodeData{handler, paramNameList}
}
