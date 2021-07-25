package httpd_test

import (
	"github.com/whoisnian/glb/httpd"
)

func pingHandler(store httpd.Store) {
	store.Respond200([]byte("pong"))
}

func sayHandler(store httpd.Store) {
	name := store.RouteParam("name")
	msg := store.RouteParam("msg")
	store.Respond200([]byte(name + " say: " + msg))
}

func anyHandler(store httpd.Store) {
	path := store.RouteParamAny()
	method := store.R.Method
	store.RespondJson(map[string]string{
		"method": method,
		"path":   path,
	})
}

func Example() {
	httpd.Handle("/test/ping", "GET", pingHandler)
	httpd.Handle("/test/say/:name/:msg", "POST", sayHandler)
	httpd.Handle("/test/any/*", "*", anyHandler)

	httpd.Start("127.0.0.1:8080")

	// Output examples:
	// curl http://127.0.0.1:8080/test/ping
	//   pong
	// curl -X POST http://127.0.0.1:8080/test/say/cat/meow
	//   cat say: meow
	// curl -X PUT http://127.0.0.1:8080/test/any/hello/world
	//   {"method":"PUT","path":"hello/world"}
}
