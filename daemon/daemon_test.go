package daemon_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/whoisnian/glb/daemon"
)

var (
	tcpPongAddr = "127.0.0.1:8000"
	tcpPongResp = []byte("pong")
)

func init() {
	daemon.Register("TestRun-daemon", noop)
	daemon.Register("TestLaunch", tcpPong)
	if daemon.Run() {
		os.Exit(0)
	}
}

func noop() {}

func tcpPong() {
	listener, err := net.Listen("tcp", tcpPongAddr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	daemon.Done()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		conn.Write(tcpPongResp)
		conn.Close()
	}
}

func tryKill(pid int) {
	if p, err := os.FindProcess(pid); err == nil {
		p.Kill()
	}
}

func TestRegister(t *testing.T) {
	defer func() { _ = recover() }()
	daemon.Register("TestRegister", noop)
	daemon.Register("TestRegister", noop)
	t.Fatal("repeated Register should panic")
}

func TestRun(t *testing.T) {
	if daemon.Run() {
		t.Fatal("should not Run without ENV_DAEMON_NAME")
	}

	flag := 0
	daemon.Register("TestRun", func() { flag = 1 })
	os.Setenv("ENV_DAEMON_NAME", "TestRun")
	os.Setenv("ENV_DAEMON_FLAG", "isDaemon")
	if !daemon.Run() || flag != 1 {
		t.Fatal("should Run handler successfully")
	}

	oldStdout := os.Stdout
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = pw

	os.Setenv("ENV_DAEMON_NAME", "TestRun-daemon")
	os.Setenv("ENV_DAEMON_FLAG", "isLauncher")
	if !daemon.Run() {
		t.Fatal("should Run launch successfully")
	}
	pw.Close()
	var data uint32
	err = binary.Read(pr, binary.LittleEndian, &data)
	if err != nil {
		t.Fatalf("launcher stdout: %v", err)
	} else if data == 0 {
		t.Fatal("Launcher should output daemon pid")
	}
	os.Stdout = oldStdout
	os.Unsetenv("ENV_DAEMON_NAME")
	os.Unsetenv("ENV_DAEMON_FLAG")
}

func TestLaunch(t *testing.T) {
	_, err := net.Dial("tcp", tcpPongAddr)
	if !errors.Is(err, syscall.ECONNREFUSED) {
		t.Fatal("daemon should not running before TestLaunch")
	}

	pid, err := daemon.Launch("TestLaunch")
	if err != nil {
		t.Fatal(err)
	}
	defer tryKill(pid)

	conn, err := net.Dial("tcp", tcpPongAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	res := make([]byte, 5)
	n, _ := conn.Read(res)
	if !bytes.Equal(tcpPongResp, res[:n]) {
		t.Fatalf("daemon response %q, want %q", res[:n], tcpPongResp)
	}
}
