// Package keeper starts daemon to create keepalive ssh connection and shares multiple sessions over it.
package keeper

import (
	"fmt"
	"os"
	"time"

	"github.com/whoisnian/glb/daemon"
	"github.com/whoisnian/glb/ssh"
	"github.com/whoisnian/glb/util/fsutil"
	xssh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

const (
	daemonName string        = "keeper"
	socketPath string        = "/tmp/keeper.socket"
	keepalive  time.Duration = 3 * time.Minute
)

func init() {
	daemon.Register(daemonName, runKeeperDaemon)
	if daemon.Run() {
		os.Exit(0)
	}
}

type Keeper struct {
	agent      *ssh.Agent
	knownhosts *ssh.Knownhosts
	keyMap     map[string]interface{}
}

func (k *Keeper) PreparePrivateKey(KeyFile string) error {
	keyPath, _ := fsutil.ResolveHomeDir(KeyFile)
	if _, ok := k.keyMap[keyPath]; ok {
		return nil
	}

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}

	privateKey, err := xssh.ParseRawPrivateKey(keyBytes)
	if err == nil {
		k.keyMap[keyPath] = privateKey
		return nil
	}

	if partial, ok := err.(*xssh.PassphraseMissingError); ok {
		if signer := k.agent.GetSigner(partial.PublicKey); signer != nil {
			k.keyMap[keyPath] = partial.PublicKey.Marshal()
			return nil
		} else {
			fmt.Printf("Enter passphrase for key '%s': ", keyPath)
			if password, err := term.ReadPassword(int(os.Stdin.Fd())); err == nil {
				if privateKey, err := xssh.ParseRawPrivateKeyWithPassphrase(keyBytes, password); err == nil {
					k.agent.AddKey(privateKey, keyPath)
					k.keyMap[keyPath] = privateKey
					return nil
				}
			}
		}
	} else {
		if pubKey, _, _, _, err := xssh.ParseAuthorizedKey(keyBytes); err == nil {
			if signer := k.agent.GetSigner(pubKey); signer != nil {
				k.keyMap[keyPath] = pubKey.Marshal()
				return nil
			}
		}
	}
	return err
}

func NewKeeper() *Keeper {
	return &Keeper{
		ssh.NewAgent(),
		ssh.NewKnownhosts(),
		make(map[string]interface{}),
	}
}
