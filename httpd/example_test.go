package httpd_test

import (
	"context"
	"fmt"
	"io"
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
	store.RespondJson(map[string]string{
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
	mux.Handle("/test/ping", "GET", pingHandler)
	mux.Handle("/test/say/:name/:msg", "POST", sayHandler)
	mux.Handle("/test/any/*", "*", anyHandler)

	server := &http.Server{Addr: ":8080", Handler: mux}
	running := make(chan struct{})
	go func() {
		close(running)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	<-running
	defer server.Shutdown(context.Background())

	printResp(http.Get("http://127.0.0.1:8080/test/ping"))
	printResp(http.Post("http://127.0.0.1:8080/test/say/cat/meow", "application/octet-stream", nil))
	printResp(http.Get("http://127.0.0.1:8080/test/any/hello/world"))

	// Output Example:
	//   pong
	//   cat say: meow
	//   {"method":"GET","path":"hello/world"}
}
