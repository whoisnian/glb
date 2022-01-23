package ssh

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/whoisnian/glb/util/fsutil"
	"github.com/whoisnian/glb/util/strutil"
	xssh "golang.org/x/crypto/ssh"
)

type Client struct {
	client *xssh.Client
}

func (c *Client) run(cmd string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = stdin
	session.Stdout = stdout
	session.Stderr = stderr

	return session.Run(cmd)
}

func (c *Client) Run(cmd string) (string, error) {
	var outbuf, errbuf bytes.Buffer
	err := c.run(cmd, nil, &outbuf, &errbuf)
	if errbuf.Len() > 0 {
		return outbuf.String(), errors.New(errbuf.String())
	}
	return outbuf.String(), err
}

func pathEscapeExceptTilde(rawPath string) string {
	if strings.HasPrefix(rawPath, "~/") {
		return "~/" + strutil.ShellEscape(rawPath[2:])
	}
	return strutil.ShellEscape(rawPath)
}

func (c *Client) GetFileWriteTo(remoteFilePath string, writer io.Writer) error {
	cmd := "cat " + pathEscapeExceptTilde(remoteFilePath)

	var errbuf bytes.Buffer
	err := c.run(cmd, nil, writer, &errbuf)
	if errbuf.Len() > 0 {
		return errors.New(errbuf.String())
	}
	return err
}

func (c *Client) PutFileReadFrom(remoteFilePath string, reader io.Reader) error {
	cmd := "tee " + pathEscapeExceptTilde(remoteFilePath)

	var errbuf bytes.Buffer
	err := c.run(cmd, reader, nil, &errbuf)
	if errbuf.Len() > 0 {
		return errors.New(errbuf.String())
	}
	return err
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
		Timeout:           10 * time.Second,
	}
	client, err := xssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}

	// ServerAliveInterval 10
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			_, _, err := client.Conn.SendRequest("keepalive@golang.org", true, nil)
			if err != nil {
				return
			}
		}
	}()

	return &Client{client}, nil
}
