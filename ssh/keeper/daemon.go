package keeper

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/whoisnian/glb/daemon"
	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/glb/ssh"
	"github.com/whoisnian/glb/util/netutil"
)

var sshClientMap *sync.Map

func runKeeperDaemon() {
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		logger.Panic(err)
	}
	defer listener.Close()
	daemon.Done()

	sshClientMap = new(sync.Map)
	wg := new(sync.WaitGroup)
	timer := time.AfterFunc(keepalive, func() {
		wg.Wait()
		listener.Close()
	})

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

			c := &keeperConn{
				conn:       conn,
				jconn:      netutil.NewJConn(conn),
				agent:      agent,
				knownhosts: knownhosts,
			}
			c.handleMsg()
			c.close()
		}()
	}
}
