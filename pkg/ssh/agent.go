package ssh

import (
	"bytes"
	"net"
	"os"

	xssh "golang.org/x/crypto/ssh"
	xagent "golang.org/x/crypto/ssh/agent"
)

func loadAgent() xagent.ExtendedAgent {
	authSockConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil
	}

	return xagent.NewClient(authSockConn)
}

func (store *Store) getSignerFromAgent(key xssh.PublicKey) xssh.Signer {
	if store.agent == nil {
		return nil
	}

	signers, err := store.agent.Signers()
	if err != nil {
		return nil
	}

	keyType, keyBytes := key.Type(), key.Marshal()
	for i := range signers {
		if keyType == signers[i].PublicKey().Type() && bytes.Equal(keyBytes, signers[i].PublicKey().Marshal()) {
			return signers[i]
		}
	}
	return nil
}

func (store *Store) addKeyToAgent(key interface{}, comment string) {
	if store.agent == nil {
		return
	}

	store.agent.Add(xagent.AddedKey{
		PrivateKey: key,
		Comment:    comment,
	})
}
