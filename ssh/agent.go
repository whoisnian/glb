package ssh

import (
	"bytes"
	"net"
	"os"

	xssh "golang.org/x/crypto/ssh"
	xagent "golang.org/x/crypto/ssh/agent"
)

type Agent struct {
	agent xagent.ExtendedAgent
}

func (a *Agent) GetSigner(key xssh.PublicKey) xssh.Signer {
	if a.agent == nil {
		return nil
	}

	signers, err := a.agent.Signers()
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

func (a *Agent) AddKey(key interface{}, comment string) {
	if a.agent == nil {
		return
	}

	a.agent.Add(xagent.AddedKey{
		PrivateKey: key,
		Comment:    comment,
	})
}

func NewAgent() *Agent {
	authSockConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return &Agent{}
	}
	return &Agent{xagent.NewClient(authSockConn)}
}
