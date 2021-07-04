package keeper

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/json"

	xssh "golang.org/x/crypto/ssh"
)

func marshalKey(key interface{}) (keyType string, keyData []byte) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		keyType = xssh.KeyAlgoRSA
		keyData, _ = json.Marshal(k)
	case *dsa.PrivateKey:
		keyType = xssh.KeyAlgoDSA
		keyData, _ = json.Marshal(k)
	case *ecdsa.PrivateKey:
		keyType = "ecdsa-sha2-xxx"
		keyData, _ = json.Marshal(k)
	case *ed25519.PrivateKey:
		keyType = xssh.KeyAlgoED25519
		keyData, _ = json.Marshal(k)
	case []byte:
		keyType = "public-key"
		keyData, _ = json.Marshal(k)
	}
	return
}

func unmarshalKey(keyType string, keyData []byte) (key interface{}) {
	switch keyType {
	case xssh.KeyAlgoRSA:
		key = new(rsa.PrivateKey)
	case xssh.KeyAlgoDSA:
		key = new(dsa.PrivateKey)
	case "ecdsa-sha2-xxx":
		key = new(ecdsa.PrivateKey)
	case xssh.KeyAlgoED25519:
		key = new(ed25519.PrivateKey)
	case "public-key":
		key = new([]byte)
	default:
		return nil
	}
	json.Unmarshal(keyData, key)
	return key
}
