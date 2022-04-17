package ssh_test

import (
	"fmt"

	"github.com/whoisnian/glb/ssh"
	"golang.org/x/term"
)

const (
	listenAddr         = "127.0.0.1:2222"
	privateKeyFile     = "~/.ssh/t_rsa"
	authorizedKeysFile = "~/.ssh/t_rsa.pub"
)

func execHandler(session *ssh.Session) {
	session.Chan.Write([]byte("exec '" + session.Command + "' successfully\n"))
	session.SendStatus(0)
}

func shellHandler(session *ssh.Session) {
	term := term.NewTerminal(session.Chan, "> ")
	for {
		line, err := term.ReadLine()
		if err != nil {
			fmt.Println("err:", err)
			break
		}
		term.Write([]byte(fmt.Sprintf("ret: %s\n", line)))
		fmt.Println("rec:", line)
	}
	term.Write([]byte("session end\n"))
}

func ExampleNewServer() {
	s := ssh.NewServer()
	s.PrepareHostKey(privateKeyFile)
	s.PrepareAuthorizedKey(authorizedKeysFile)

	s.Handle("exec", execHandler)
	s.Handle("shell", shellHandler)

	s.ListenAndServe(listenAddr, func(e error) { fmt.Println(e) })
}
