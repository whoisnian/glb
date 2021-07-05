package keeper

import (
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/whoisnian/glb/ssh"
	"github.com/whoisnian/glb/util/netutil"
	xssh "golang.org/x/crypto/ssh"
)

type keeperConn struct {
	conn       net.Conn
	jconn      *netutil.JConn
	agent      *ssh.Agent
	knownhosts *ssh.Knownhosts
	sshClient  *xssh.Client
}

func (c *keeperConn) close() {
	c.jconn.Close()
	c.conn.Close()
}

type keeperMsg struct {
	Action string
	Data   json.RawMessage
}

type keeperRes struct {
	Status int64
	Result string
}

func (c *keeperConn) handleMsg() {
	msg := new(keeperMsg)
	var res keeperRes
	for c.jconn.Accept(msg) {
		switch msg.Action {
		case "create-client":
			res = c.createClient(msg.Data)
		case "run-command":
			res = c.runCommand(msg.Data)
		default:
			res = keeperRes{404, "unknown msg.Action"}
		}
		c.jconn.Send(res)
	}
}

type createClientData struct {
	Addr    string
	User    string
	KeyFile string
	KeyType string
	KeyData json.RawMessage
}

func (d createClientData) tag() string {
	return strings.Join([]string{d.Addr, d.User, d.KeyFile}, "|")
}

func (c *keeperConn) createClient(data json.RawMessage) keeperRes {
	var d createClientData
	json.Unmarshal(data, &d)

	if sshClient, ok := sshClientMap.Load(d.tag()); ok {
		c.sshClient = sshClient.(*xssh.Client)
		return keeperRes{200, "reuse existing sshClient"}
	}

	var signer xssh.Signer
	var err error
	key := unmarshalKey(d.KeyType, d.KeyData)
	if d.KeyType == "public-key" {
		publicKey, err := xssh.ParsePublicKey(*key.(*[]byte))
		if err != nil {
			return keeperRes{401, err.Error()}
		}
		signer = c.agent.GetSigner(publicKey)
		if signer == nil {
			return keeperRes{401, d.KeyFile + " is passphrase protected"}
		}
	} else {
		signer, err = xssh.NewSignerFromKey(key)
		if err != nil {
			return keeperRes{401, err.Error()}
		}
	}
	authMethod := xssh.PublicKeys(signer)

	config := &xssh.ClientConfig{
		User:              d.User,
		Auth:              []xssh.AuthMethod{authMethod},
		HostKeyCallback:   c.knownhosts.AcceptNewHostKeyCallback,
		HostKeyAlgorithms: c.knownhosts.OrderedHostKeyAlgorithms(d.Addr),
		Timeout:           10 * time.Second,
	}
	c.sshClient, err = xssh.Dial("tcp", d.Addr, config)
	if err != nil {
		return keeperRes{401, err.Error()}
	}

	// ServerAliveInterval 10
	go func() {
		defer sshClientMap.Delete(d.tag())
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			_, _, err := c.sshClient.Conn.SendRequest("keepalive@golang.org", true, nil)
			if err != nil {
				return
			}
		}
	}()

	sshClientMap.Store(d.tag(), c.sshClient)
	return keeperRes{200, "create new sshClient"}
}

type runCommandData struct {
	Cmd string
}

func (c *keeperConn) runCommand(data json.RawMessage) keeperRes {
	var d runCommandData
	json.Unmarshal(data, &d)

	session, err := c.sshClient.NewSession()
	if err != nil {
		return keeperRes{500, err.Error()}
	}
	defer session.Close()

	out, err := session.Output(d.Cmd)
	if err != nil {
		return keeperRes{500, err.Error()}
	}
	return keeperRes{200, string(out)}
}
