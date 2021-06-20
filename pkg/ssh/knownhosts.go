package ssh

import (
	"net"

	"github.com/whoisnian/glb/pkg/fs"
	xssh "golang.org/x/crypto/ssh"
	xknownhosts "golang.org/x/crypto/ssh/knownhosts"
)

var supportedHostKeyAlgos = []string{
	xssh.CertAlgoRSAv01, xssh.CertAlgoDSAv01, xssh.CertAlgoECDSA256v01,
	xssh.CertAlgoECDSA384v01, xssh.CertAlgoECDSA521v01, xssh.CertAlgoED25519v01,
	xssh.KeyAlgoECDSA256, xssh.KeyAlgoECDSA384, xssh.KeyAlgoECDSA521,
	xssh.KeyAlgoRSA, xssh.KeyAlgoDSA,
	xssh.KeyAlgoED25519,
}

type emptyPublicKey struct{ xssh.PublicKey }

func (emptyPublicKey) Type() string                         { return "" }
func (emptyPublicKey) Marshal() []byte                      { return nil }
func (emptyPublicKey) Verify([]byte, *xssh.Signature) error { return nil }

func loadKnownhosts() xssh.HostKeyCallback {
	hostKeyPath, _ := fs.Clean("~/.ssh/known_hosts")
	hostKeyCallback, err := xknownhosts.New(hostKeyPath)
	if err != nil {
		return nil
	}
	return hostKeyCallback
}

func (store *Store) OrderedHostKeyAlgorithms(addr string) []string {
	if store.hostKeyCheck == nil {
		return supportedHostKeyAlgos
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return supportedHostKeyAlgos
	}
	err = store.hostKeyCheck(addr, tcpAddr, emptyPublicKey{})
	if err, ok := err.(*xknownhosts.KeyError); ok && len(err.Want) > 0 {
		has := make(map[string]bool)
		for i := range err.Want {
			has[err.Want[i].Key.Type()] = true
		}
		algos := make([]string, len(supportedHostKeyAlgos))
		copy(algos, supportedHostKeyAlgos)
		pos := 0
		for cur := range algos {
			if has[algos[cur]] {
				algos[pos], algos[cur] = algos[cur], algos[pos]
				pos++
			}
		}
		return algos
	}
	return supportedHostKeyAlgos
}

func (store *Store) AcceptNewHostKeyCallback(hostname string, remote net.Addr, key xssh.PublicKey) error {
	if store.hostKeyCheck == nil {
		return nil
	}

	err := store.hostKeyCheck(hostname, remote, key)
	if err, ok := err.(*xknownhosts.KeyError); ok && len(err.Want) == 0 {
		return nil
	}
	return err
}
