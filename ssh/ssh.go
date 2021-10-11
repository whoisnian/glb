// Package ssh implements an SSH client(autoload SSH_AUTH_SOCK and knownhosts, only allow public key authentication).
package ssh

import (
	"fmt"
	"os"

	"github.com/whoisnian/glb/util/fsutil"
	xssh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type Store struct {
	agent      *Agent
	knownhosts *Knownhosts
	signerMap  map[string]xssh.Signer
}

func (store *Store) PreparePrivateKey(KeyFile string) error {
	keyPath, _ := fsutil.ResolveHomeDir(KeyFile)
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
		if signer = store.agent.GetSigner(partial.PublicKey); signer != nil {
			store.signerMap[keyPath] = signer
			return nil
		} else {
			fmt.Printf("Enter passphrase for key '%s': ", keyPath)
			if password, err := term.ReadPassword(int(os.Stdin.Fd())); err == nil {
				if privateKey, err := xssh.ParseRawPrivateKeyWithPassphrase(keyBytes, password); err == nil {
					if signer, err = xssh.NewSignerFromKey(privateKey); err == nil {
						store.agent.AddKey(privateKey, keyPath)
						store.signerMap[keyPath] = signer
						return nil
					}
				}
			}
		}
	} else {
		if pubKey, _, _, _, err := xssh.ParseAuthorizedKey(keyBytes); err == nil {
			if signer := store.agent.GetSigner(pubKey); signer != nil {
				store.signerMap[keyPath] = signer
				return nil
			}
		}
	}
	return err
}

func NewStore() *Store {
	return &Store{
		NewAgent(),
		NewKnownhosts(),
		make(map[string]xssh.Signer),
	}
}
