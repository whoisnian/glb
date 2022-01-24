package keeper

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"syscall"
	"time"

	"github.com/whoisnian/glb/daemon"
	"github.com/whoisnian/glb/util/fsutil"
	"github.com/whoisnian/glb/util/netutil"
	"github.com/whoisnian/glb/util/strutil"
)

type Client struct {
	conn  net.Conn
	jconn *netutil.JConn
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) Result() (res keeperRes) {
	c.jconn.Accept(&res)
	return res
}

func (c *Client) Run(cmd string) (string, error) {
	data, _ := json.Marshal(runCommandData{cmd, nil})
	c.jconn.Send(keeperMsg{
		Action: "run-command",
		Data:   data,
	})
	res := c.Result()
	if res.Status == 500 {
		return "", errors.New(res.Result)
	}
	var d runCommandRes
	err := json.Unmarshal(res.Data, &d)
	if err != nil {
		return "", err
	}
	if res.Status != 200 {
		if len(d.Stderr) > 0 {
			return "", errors.New(string(d.Stderr))
		}
		return "", errors.New(res.Result)
	}
	return string(d.Stdout), nil
}

func (c *Client) GetFileWriteTo(remoteFilePath string, writer io.Writer) error {
	cmd := "cat " + strutil.ShellEscapeExceptTilde(remoteFilePath)

	data, _ := json.Marshal(runCommandData{cmd, nil})
	c.jconn.Send(keeperMsg{
		Action: "run-command",
		Data:   data,
	})
	res := c.Result()
	if res.Status == 500 {
		return errors.New(res.Result)
	}
	var d runCommandRes
	err := json.Unmarshal(res.Data, &d)
	if err != nil {
		return err
	}
	if res.Status != 200 {
		if len(d.Stderr) > 0 {
			return errors.New(string(d.Stderr))
		}
		return errors.New(res.Result)
	}
	_, err = writer.Write(d.Stdout)
	return err
}

func (c *Client) PutFileReadFrom(remoteFilePath string, reader io.Reader) error {
	cmd := "tee " + strutil.ShellEscapeExceptTilde(remoteFilePath)
	stdin, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	data, _ := json.Marshal(runCommandData{cmd, stdin})
	c.jconn.Send(keeperMsg{
		Action: "run-command",
		Data:   data,
	})
	res := c.Result()
	if res.Status == 500 {
		return errors.New(res.Result)
	}
	var d runCommandRes
	err = json.Unmarshal(res.Data, &d)
	if err != nil {
		return err
	}
	if res.Status != 200 {
		if len(d.Stderr) > 0 {
			return errors.New(string(d.Stderr))
		}
		return errors.New(res.Result)
	}
	return nil
}

func (k *Keeper) NewClient(addr string, user string, keyFile string) (*Client, error) {
	conn, err := net.DialTimeout("unix", socketPath, 10*time.Second)
	if err != nil {
		if errors.Is(err, syscall.ENOENT) {
			daemon.Launch(daemonName)

			conn, err = net.DialTimeout("unix", socketPath, 10*time.Second)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	client := &Client{conn: conn}
	if client.jconn, err = netutil.NewJConn(conn); err != nil {
		return nil, err
	}
	keyPath, _ := fsutil.ResolveHomeDir(keyFile)
	key := k.keyMap[keyPath]

	keyType, keyData := marshalKey(key)
	data, _ := json.Marshal(createClientData{
		Addr:    addr,
		User:    user,
		KeyFile: keyFile,
		KeyType: keyType,
		KeyData: keyData,
	})
	client.jconn.Send(keeperMsg{
		Action: "create-client",
		Data:   data,
	})
	if res := client.Result(); res.Status != 200 {
		return nil, errors.New(res.Result)
	}
	return client, nil
}
