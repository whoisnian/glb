//go:build !windows

// Package daemon creates orphan process as daemon.
//  1. Current process use Launch() to start launcher and wait.
//  2. Launcher use launch() to start daemon and wait.
//  3. Daemon use Done() to kill its launcher and continue.
//  4. Launcher exits and current process continue.
package daemon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
)

const (
	envDaemonName = "ENV_DAEMON_NAME"
	envDaemonFlag = "ENV_DAEMON_FLAG"
)

var handlerMap = make(map[string]func())

// Register registers the handler function for the given name. Usually used in package init function.
func Register(name string, handler func()) {
	if _, ok := handlerMap[name]; ok {
		panic("handler '" + name + "' already registered")
	}
	handlerMap[name] = handler
}

// Run starts launcher or daemon based on environment variables. Usually used in package init function.
//
// Example:
//
//	func init() {
//	    daemon.Register(daemonName, runDaemon)
//	    if daemon.Run() {
//	        os.Exit(0)
//	    }
//	}
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

// Launch should be invoked in the main program process to start launcher as daemon's parent.
func Launch(name string) (pid int, err error) {
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), envDaemonName+"="+name, envDaemonFlag+"=isLauncher")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		return 0, fmt.Errorf("start launcher: %v", err)
	} else if stderr.Len() > 0 {
		return 0, errors.New(stderr.String())
	} else {
		var data uint32
		if err = binary.Read(&stdout, binary.LittleEndian, &data); err != nil {
			return 0, fmt.Errorf("launcher stdout: %v", err)
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
