package httpd_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/whoisnian/glb/httpd"
)

func pingHandler(store *httpd.Store) {
	store.Respond200([]byte("pong"))
}

func sayHandler(store *httpd.Store) {
	name := store.RouteParam("name")
	msg := store.RouteParam("msg")
	store.Respond200([]byte(name + " say: " + msg))
}

func anyHandler(store *httpd.Store) {
	path := store.RouteParamAny()
	method := store.R.Method
	store.RespondJson(http.StatusOK, map[string]string{
		"method": method,
		"path":   path,
	})
}

func printResp(resp *http.Response, err error) {
	if err != nil {
		fmt.Println(err)
	} else {
		data, _ := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		fmt.Println(string(data))
	}
}

func Example() {
	mux := httpd.NewMux()
	mux.Handle("/test/ping", http.MethodGet, pingHandler)
	mux.Handle("/test/say/:name/:msg", http.MethodPost, sayHandler)
	mux.Handle("/test/any/*", httpd.MethodAll, anyHandler)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	server := &http.Server{Addr: ln.Addr().String(), Handler: mux}
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	defer server.Shutdown(context.Background())

	printResp(http.Get("http://" + server.Addr + "/test/ping"))
	printResp(http.Post("http://"+server.Addr+"/test/say/cat/meow", "application/octet-stream", nil))
	printResp(http.Get("http://" + server.Addr + "/test/any/hello/world"))

	// Output:
	// pong
	// cat say: meow
	// {"method":"GET","path":"hello/world"}
}
