package ssh

import (
	"github.com/whoisnian/glb/util/fsutil"
	xssh "golang.org/x/crypto/ssh"
)

type Client struct {
	client *xssh.Client
}

func (c *Client) Run(cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	out, err := session.Output(cmd)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (c *Client) Close() {
	if c != nil && c.client != nil {
		c.client.Close()
	}
}

func (store *Store) NewClient(addr string, user string, keyFile string) (*Client, error) {
	keyPath, _ := fsutil.ResolveHomeDir(keyFile)
	authMethod := xssh.PublicKeys(store.signerMap[keyPath])

	config := &xssh.ClientConfig{
		User:              user,
		Auth:              []xssh.AuthMethod{authMethod},
		HostKeyCallback:   store.knownhosts.AcceptNewHostKeyCallback,
		HostKeyAlgorithms: store.knownhosts.OrderedHostKeyAlgorithms(addr),
	}
	client, err := xssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}
