package keeper

import (
	"bytes"
	"encoding/json"
	"io"
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
	Data   json.RawMessage
}

func (c *keeperConn) handleMsg() (err error) {
	msg := new(keeperMsg)
	var res keeperRes
	for {
		if err = c.jconn.Accept(msg); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch msg.Action {
		case "create-client":
			res = c.createClient(msg.Data)
		case "run-command":
			res = c.runCommand(msg.Data)
		default:
			res = keeperRes{404, "unknown msg.Action", nil}
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
		return keeperRes{200, "reuse existing sshClient", nil}
	}

	var signer xssh.Signer
	var err error
	key := unmarshalKey(d.KeyType, d.KeyData)
	if d.KeyType == "public-key" {
		publicKey, err := xssh.ParsePublicKey(*key.(*[]byte))
		if err != nil {
			return keeperRes{401, err.Error(), nil}
		}
		signer = c.agent.GetSigner(publicKey)
		if signer == nil {
			return keeperRes{401, d.KeyFile + " is passphrase protected", nil}
		}
	} else {
		signer, err = xssh.NewSignerFromKey(key)
		if err != nil {
			return keeperRes{401, err.Error(), nil}
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
		return keeperRes{401, err.Error(), nil}
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
	return keeperRes{200, "create new sshClient", nil}
}

type runCommandData struct {
	Cmd string
}

func (c *keeperConn) runCommand(data json.RawMessage) keeperRes {
	var d runCommandData
	json.Unmarshal(data, &d)

	var outbuf, errbuf bytes.Buffer
	err := c.run(d.Cmd, nil, &outbuf, &errbuf)
	if errbuf.Len() > 0 {
		return keeperRes{500, errbuf.String(), nil}
	} else if err != nil {
		return keeperRes{500, err.Error(), nil}
	}

	return keeperRes{200, outbuf.String(), nil}
}

func (c *keeperConn) run(cmd string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = stdin
	session.Stdout = stdout
	session.Stderr = stderr

	return session.Run(cmd)
}
