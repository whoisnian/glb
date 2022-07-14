// Package daemon create orphan process as daemon.
package daemon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"os/exec"
	"os/signal"
)

const (
	envDaemonName = "ENV_DAEMON_NAME"
	envDaemonFlag = "ENV_DAEMON_FLAG"
)

var handlerMap = make(map[string]func())

func Register(name string, handler func()) {
	if _, ok := handlerMap[name]; ok {
		panic("handler '" + name + "' already registered")
	}
	handlerMap[name] = handler
}

func Run() bool {
	if name, ok := os.LookupEnv(envDaemonName); ok {
		if handler, ok := handlerMap[name]; ok {
			switch os.Getenv(envDaemonFlag) {
			case "isLauncher":
				launch(name)
			case "isDaemon":
				handler()
			}
		}
		return true
	} else {
		return false
	}
}

func launch(name string) {
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), envDaemonName+"="+name, envDaemonFlag+"=isDaemon")
	if err := cmd.Start(); err != nil {
		os.Stderr.Write([]byte("start daemon: " + err.Error()))
		return
	} else {
		binary.Write(os.Stdout, binary.LittleEndian, uint32(cmd.Process.Pid))
	}

	finished := make(chan struct{})
	go func() {
		if err := cmd.Wait(); err != nil {
			os.Stderr.Write([]byte("daemon: " + err.Error()))
		}
		close(finished)
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)
	select {
	case <-finished:
	case <-interrupt:
	}
}

// Launch start launcher as daemon's parent process.
func Launch(name string) (pid int, err error) {
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), envDaemonName+"="+name, envDaemonFlag+"=isLauncher")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		return 0, errors.New("start launcher: " + err.Error())
	} else if stderr.Len() > 0 {
		return 0, errors.New(stderr.String())
	} else {
		var data uint32
		if err = binary.Read(&stdout, binary.LittleEndian, &data); err != nil {
			return 0, errors.New("launcher stdout: " + err.Error())
		}
		return int(data), nil
	}
}

// Done should be invoked in daemon to kill its launcher.
func Done() (err error) {
	var p *os.Process
	if p, err = os.FindProcess(os.Getppid()); err == nil {
		if err = p.Signal(os.Interrupt); err == nil {
			return nil
		}
	}
	return err
}
