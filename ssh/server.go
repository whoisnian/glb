package ssh

import (
	"errors"
	"net"
	"os"
	"sync"
	"time"

	"github.com/whoisnian/glb/util/fsutil"
	"golang.org/x/crypto/ssh"
)

type Server struct {
	serverConfig  *ssh.ServerConfig
	authorizedMap map[string]bool
	handlerMap    map[string]func(*Session)
}

func (S *Server) PrepareHostKey(keyFile string) error {
	keyPath, _ := fsutil.ResolveHomeDir(keyFile)
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return err
	}
	S.serverConfig.AddHostKey(signer)
	return nil
}

func (S *Server) PrepareAuthorizedKey(KeyFile string) error {
	keyPath, _ := fsutil.ResolveHomeDir(KeyFile)
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	for len(keyBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(keyBytes)
		if err != nil {
			return err
		}
		S.authorizedMap[string(pubKey.Marshal())] = true
		keyBytes = rest
	}
	return nil
}

func (S *Server) PublicKeyCallback(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
	if S.authorizedMap[string(pubKey.Marshal())] {
		return &ssh.Permissions{
			Extensions: map[string]string{"pubkey-fp": ssh.FingerprintSHA256(pubKey)},
		}, nil
	}
	return nil, errors.New("unknown public key")
}

func NewServer() *Server {
	s := &Server{}
	s.serverConfig = &ssh.ServerConfig{PublicKeyCallback: s.PublicKeyCallback}
	s.authorizedMap = make(map[string]bool)
	s.handlerMap = make(map[string]func(*Session))
	return s
}

func (S *Server) Handle(reqType string, handler func(*Session)) {
	S.handlerMap[reqType] = handler
}

type Session struct {
	Chan    ssh.Channel
	Reqs    <-chan *ssh.Request
	wg      sync.WaitGroup
	once    sync.Once
	Type    string // "shell" or "exec"
	Command string // command to run if type is exec
}

func (session *Session) wgDoneOnce() {
	session.once.Do(func() { session.wg.Done() })
}

func (session *Session) SendStatus(status uint32) error {
	msg := struct{ Status uint32 }{status}
	_, err := session.Chan.SendRequest("exit-status", false, ssh.Marshal(&msg))
	return err
}

func (S *Server) ListenAndServe(listenAddr string, errorCallback func(error)) error {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			_, chans, reqs, err := ssh.NewServerConn(conn, S.serverConfig)
			if err != nil && errorCallback != nil {
				errorCallback(err)
				return
			}
			go ssh.DiscardRequests(reqs)
			for newChannel := range chans {
				if newChannel.ChannelType() != "session" {
					newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
					continue
				}
				channel, requests, err := newChannel.Accept()
				if err != nil {
					errorCallback(err)
					continue
				}
				session := &Session{Chan: channel, Reqs: requests}
				session.wg.Add(1)
				time.AfterFunc(time.Second*5, session.wgDoneOnce)
				go func() {
					for req := range session.Reqs {
						if req.Type == "exec" {
							c := struct{ Cmd string }{}
							ssh.Unmarshal(req.Payload, &c)
							session.Command = c.Cmd
							session.Type = req.Type
							session.wgDoneOnce()
						} else if req.Type == "shell" {
							session.Type = req.Type
							session.wgDoneOnce()
						}
						if req.WantReply {
							req.Reply(req.Type == "shell" || req.Type == "pty-req" || req.Type == "exec", nil)
						}
					}
				}()
				go S.serve(session)
			}
		}()
	}
}

func (S *Server) serve(session *Session) {
	defer session.Chan.Close()

	session.wg.Wait()

	if handler, ok := S.handlerMap[session.Type]; ok {
		handler(session)
	}
}
