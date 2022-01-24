package ssh

import (
	"net"

	"github.com/whoisnian/glb/util/fsutil"
	xssh "golang.org/x/crypto/ssh"
	xknownhosts "golang.org/x/crypto/ssh/knownhosts"
)

// from: https://cs.opensource.google/go/x/crypto/+/master:ssh/common.go;l=71
var supportedHostKeyAlgos = []string{
	xssh.CertSigAlgoRSASHA2512v01, xssh.CertSigAlgoRSASHA2256v01,
	xssh.CertSigAlgoRSAv01, xssh.CertAlgoDSAv01, xssh.CertAlgoECDSA256v01,
	xssh.CertAlgoECDSA384v01, xssh.CertAlgoECDSA521v01, xssh.CertAlgoED25519v01,
	xssh.KeyAlgoECDSA256, xssh.KeyAlgoECDSA384, xssh.KeyAlgoECDSA521,
	xssh.SigAlgoRSASHA2512, xssh.SigAlgoRSASHA2256,
	xssh.SigAlgoRSA, xssh.KeyAlgoDSA,
	xssh.KeyAlgoED25519,
}

type emptyPublicKey struct{ xssh.PublicKey }

func (emptyPublicKey) Type() string                         { return "" }
func (emptyPublicKey) Marshal() []byte                      { return nil }
func (emptyPublicKey) Verify([]byte, *xssh.Signature) error { return nil }

type Knownhosts struct {
	hostKeyCheck xssh.HostKeyCallback
}

func (k *Knownhosts) HostKeyCallback() xssh.HostKeyCallback {
	return k.hostKeyCheck
}

func (k *Knownhosts) AcceptNewHostKeyCallback(hostname string, remote net.Addr, key xssh.PublicKey) error {
	if k.hostKeyCheck == nil {
		return nil
	}

	err := k.hostKeyCheck(hostname, remote, key)
	if err, ok := err.(*xknownhosts.KeyError); ok && len(err.Want) == 0 {
		return nil
	}
	return err
}

func (k *Knownhosts) OrderedHostKeyAlgorithms(addr string) []string {
	if k.hostKeyCheck == nil {
		return supportedHostKeyAlgos
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return supportedHostKeyAlgos
	}
	err = k.hostKeyCheck(addr, tcpAddr, emptyPublicKey{})
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

func NewKnownhosts() *Knownhosts {
	hostKeyPath, _ := fsutil.ResolveHomeDir("~/.ssh/known_hosts")
	hostKeyCallback, err := xknownhosts.New(hostKeyPath)
	if err != nil {
		return &Knownhosts{}
	}
	return &Knownhosts{hostKeyCallback}
}
