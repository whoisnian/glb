package keeper

import (
	"encoding/json"
	"errors"
	"net"
	"syscall"
	"time"

	"github.com/whoisnian/glb/daemon"
	"github.com/whoisnian/glb/util/fsutil"
	"github.com/whoisnian/glb/util/netutil"
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
	data, _ := json.Marshal(runCommandData{cmd})
	c.jconn.Send(keeperMsg{
		Action: "run-command",
		Data:   data,
	})
	if res := c.Result(); res.Status != 200 {
		return "", errors.New(res.Result)
	} else {
		return res.Result, nil
	}
}

// func (c *Client) GetFileWriteTo(remoteFilePath string, writer io.Writer) error {
// 	cmd := "cat " + pathEscapeExceptTilde(remoteFilePath)

// 	var errbuf bytes.Buffer
// 	err := c.run(cmd, nil, writer, &errbuf)
// 	if errbuf.Len() > 0 {
// 		return errors.New(errbuf.String())
// 	}
// 	return err
// }

// func (c *Client) PutFileReadFrom(remoteFilePath string, reader io.Reader) error {
// 	cmd := "tee " + pathEscapeExceptTilde(remoteFilePath)

// 	var errbuf bytes.Buffer
// 	err := c.run(cmd, reader, nil, &errbuf)
// 	if errbuf.Len() > 0 {
// 		return errors.New(errbuf.String())
// 	}
// 	return err
// }

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
