package httpd

import (
	"errors"
	"reflect"
	"runtime"
	"slices"
)

const (
	routeParam    string = "/:param"
	routeParamAny string = "/:any"
)

// RouteInfo exports route information for logs and metrics.
type RouteInfo struct {
	Path        string
	Method      string
	HandlerName string
	HandlerFunc HandlerFunc
	Middlewares *[]HandlerFunc
}

func nameOfFunc(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func newRouteInfo(path string, method string, handler HandlerFunc, middlewares *[]HandlerFunc) *RouteInfo {
	return &RouteInfo{
		Path:        path,
		Method:      method,
		HandlerName: nameOfFunc(handler),
		HandlerFunc: handler,
		Middlewares: middlewares,
	}
}

type treeNode struct {
	next          map[string]*treeNode
	info          *RouteInfo
	paramNameList []string
}

func (node *treeNode) nextNodeOrNew(name string) (resNode *treeNode) {
	// It's ok to retrieve an element from nil map.
	// But it will panic if insert an element to nil map.
	if resNode, ok := node.next[name]; ok {
		return resNode
	}
	if node.next == nil {
		node.next = make(map[string]*treeNode)
	}
	node.next[name] = new(treeNode)
	return node.next[name]
}

func (node *treeNode) methodNodeOrNil(method string) (resNode *treeNode) {
	if resNode, ok := node.next[methodTagMap[method]]; ok {
		return resNode
	}
	return node.next[methodTagMap[MethodAll]]
}

func parseRoute(node *treeNode, path string, method string, info *RouteInfo) (paramsCnt int, err error) {
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
			// skip empty fragment
		} else if path[left+1:right] == "*" {
			paramNameList = append(paramNameList, routeParamAny)
			node = node.nextNodeOrNew(routeParamAny)
			break
		} else if path[left+1] == ':' {
			paramName := path[left+2 : right]
			if paramName == "" || slices.Contains(paramNameList, paramName) {
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
	node.info = info
	node.paramNameList = paramNameList
	return len(paramNameList), nil
}

// about trailing slash:
//
//	`/foo/bar`  will be matched by `/foo/bar`
//	`/foo/bar/` will be matched by `/foo/bar/:param` or `/foo/bar/*`
//	`/`         will be matched by `/` first and then `/:param` or `/*`
func findRoute(node *treeNode, path string, method string, params *Params) (info *RouteInfo) {
	var length, left, right int = len(path), 0, 0
	if length == 1 {
		if n := node.methodNodeOrNil(method); n != nil {
			// if `/` is matched by `/`, skip `/:param` and `/*`
			params.K = n.paramNameList
			return n.info
		}
	}
	for ; right <= length; right++ {
		if right < length && path[right] != '/' {
			continue
		}
		if right-left < 2 && right < length { // check routeParam if current is last fragment
			// skip empty fragment
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
		return node.info
	} else {
		return nil
	}
}
