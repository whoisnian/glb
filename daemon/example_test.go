package daemon_test

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/whoisnian/glb/daemon"
)

const daemonName string = "pongServer"
const listenAddr string = "127.0.0.1:8000"

func init() {
	daemon.Register(daemonName, runDaemon)
	if daemon.Run() {
		os.Exit(0)
	}
}

func runDaemon() {
	listener, err := net.Listen("tcp", listenAddr)
	daemon.Done()
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go func(c net.Conn) {
			c.Write([]byte("pong"))
			c.Close()
		}(conn)
	}
}

func Example() {
	conn, err := net.Dial("tcp", listenAddr)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			daemon.Launch(daemonName)
			if conn, err = net.Dial("tcp", listenAddr); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	defer conn.Close()

	res := make([]byte, 5)
	n, _ := conn.Read(res)
	fmt.Println(string(res[:n]))

	// Output:
	// pong
}
