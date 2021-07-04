// Package daemon create orphan process as daemon.
package daemon

import (
	"os"
	"os/exec"
	"os/signal"
)

const launcherEnv string = "LAUNCHER_NAME"
const launcherIng string = "LAUNCHER_ING"

var handlerMap = make(map[string]func())

func Register(name string, handler func()) {
	if _, ok := handlerMap[name]; ok {
		panic("Handler '" + name + "' already registered")
	}
	handlerMap[name] = handler
}

func Run() bool {
	name := os.Getenv(launcherEnv)
	if handler, ok := handlerMap[name]; ok {
		if os.Getenv(launcherIng) == "true" {
			launch(name)
		} else {
			handler()
		}
		return true
	} else {
		return false
	}
}

func launch(name string) {
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), launcherEnv+"="+name, launcherIng+"=false")
	cmd.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}

func Launch(name string) {
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), launcherEnv+"="+name, launcherIng+"=true")
	cmd.Run()
}

// Done should be invoked in handler to kill its launcher.
func Done() (err error) {
	var p *os.Process
	if p, err = os.FindProcess(os.Getppid()); err == nil {
		if err = p.Signal(os.Interrupt); err == nil {
			return nil
		}
	}
	return err
}
