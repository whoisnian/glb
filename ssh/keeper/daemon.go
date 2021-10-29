package keeper

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/whoisnian/glb/daemon"
	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/glb/ssh"
	"github.com/whoisnian/glb/util/netutil"
)

var sshClientMap *sync.Map

func runKeeperDaemon() {
	listener, err := net.Listen("unix", socketPath)
	daemon.Done()
	if err != nil {
		logger.Panic(err)
	}
	defer listener.Close()

	sshClientMap = new(sync.Map)
	wg := new(sync.WaitGroup)
	timer := time.AfterFunc(keepalive, func() {
		wg.Wait()
		listener.Close()
	})

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-interrupt
		listener.Close()
	}()

	agent := ssh.NewAgent()
	knownhosts := ssh.NewKnownhosts()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			continue
		}
		timer.Reset(keepalive)

		wg.Add(1)
		go func() {
			defer wg.Done()

			jconn, err := netutil.NewJConn(conn)
			if err != nil {
				logger.Error(err)
				conn.Close()
				return
			}

			c := &keeperConn{
				conn:       conn,
				jconn:      jconn,
				agent:      agent,
				knownhosts: knownhosts,
			}
			defer c.close()
			if err = c.handleMsg(); err != nil {
				logger.Error(err)
			}
		}()
	}
}
