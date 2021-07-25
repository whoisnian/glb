// Package httpd implements a simple HTTP router with path parameters support.
package httpd

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/whoisnian/glb/logger"
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
	handler       Handler
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
				logger.Fatal("Invalid fragment '", fragments[i], "' in routePath: '", path, "'")
			}
			paramNameList = append(paramNameList, paramName)
			node = node.nextNodeOrNew(routeParam)
		} else {
			node = node.nextNodeOrNew(fragments[i])
		}
	}

	methodTag, ok := methodList[method]
	if !ok {
		logger.Fatal("Invalid method '", method, "' for routePath: '", path, "'")
	}
	if _, ok = node.next[methodTag]; ok {
		logger.Fatal("Duplicate method '", method, "' for routePath: '", path, "'")
	}
	return node.nextNodeOrNew(methodTag), paramNameList
}

func findRoute(node *routeNode, path string, method string) (*routeNode, []string) {
	var paramValueList []string
	fragments := strings.Split(path, "/")
	for i := range fragments {
		if len(fragments[i]) < 1 {
			continue
		} else if res, ok := node.next[fragments[i]]; ok {
			node = res
		} else if res, ok := node.next[routeParam]; ok {
			value, err := url.PathUnescape(fragments[i])
			if err != nil {
				logger.Error("Invalid fragment '", fragments[i], "' in routePath: '", path, "'")
				return nil, nil
			}
			paramValueList = append(paramValueList, value)
			node = res
		} else if res, ok := node.next[routeParamAny]; ok {
			value, err := url.PathUnescape(strings.Join(fragments[i:], "/"))
			if err != nil {
				logger.Error("Invalid fragment '", fragments[i], "' in routePath: '", path, "'")
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

type serveMux struct {
	mu   sync.RWMutex
	root *routeNode
}

func (mux *serveMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	store := Store{&statusResponseWriter{w, http.StatusOK}, r, make(map[string]string)}

	defer func() {
		if err := recover(); err != nil {
			store.Error500("Internal Server Error")
		}

		logger.Req(
			r.RemoteAddr[0:strings.IndexByte(r.RemoteAddr, ':')], " [",
			store.W.status, "] ",
			r.Method, " ",
			strconv.Quote(r.URL.Path), " ",
			r.UserAgent(), " ",
			time.Since(start).Milliseconds(),
		)
	}()

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

var mux *serveMux

func init() {
	mux = new(serveMux)
	mux.root = new(routeNode)
}

// Handle registers the handler for the given routePath and method.
func Handle(path string, method string, handler Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	node, paramNameList := parseRoute(mux.root, path, method)
	node.data = &nodeData{handler, paramNameList}
}

// Start listens on the addr and then creates goroutine to handle each request.
func Start(addr string) {
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Fatal(err)
	}
}
