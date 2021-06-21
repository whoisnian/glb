// Package ssh implements an SSH client(autoload SSH_AUTH_SOCK and knownhosts, only allow public key authentication).
package ssh

import (
	"fmt"
	"os"

	"github.com/whoisnian/glb/fs"
	xssh "golang.org/x/crypto/ssh"
	xagent "golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

type Store struct {
	agent        xagent.ExtendedAgent
	hostKeyCheck xssh.HostKeyCallback
	signerMap    map[string]xssh.Signer
}

func (store *Store) PreparePrivateKey(KeyFile string) error {
	keyPath, _ := fs.Clean(KeyFile)
	if _, ok := store.signerMap[keyPath]; ok {
		return nil
	}

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}

	signer, err := xssh.ParsePrivateKey(keyBytes)
	if err == nil {
		store.signerMap[keyPath] = signer
		return nil
	}

	if partial, ok := err.(*xssh.PassphraseMissingError); ok {
		if signer = store.getSignerFromAgent(partial.PublicKey); signer != nil {
			store.signerMap[keyPath] = signer
			return nil
		} else {
			fmt.Printf("Enter passphrase for key '%s': ", keyPath)
			if password, err := term.ReadPassword(int(os.Stdin.Fd())); err == nil {
				if privateKey, err := xssh.ParseRawPrivateKeyWithPassphrase(keyBytes, password); err == nil {
					if signer, err = xssh.NewSignerFromKey(privateKey); err == nil {
						store.addKeyToAgent(privateKey, keyPath)
						store.signerMap[keyPath] = signer
						return nil
					}
				}
			}
		}
	}
	return err
}

func NewStore() *Store {
	return &Store{
		loadAgent(),
		loadKnownhosts(),
		make(map[string]xssh.Signer),
	}
}
