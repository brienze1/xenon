package mock

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"

	gossh "golang.org/x/crypto/ssh"
)

type SshKey struct {
	Signer     gossh.Signer
	Private    ed25519.PrivateKey
	PrivatePem []byte
}

func GenerateSshKey(passphrase string) (*SshKey, error) {
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	signer, err := gossh.NewSignerFromKey(private)
	if err != nil {
		return nil, err
	}
	var block *pem.Block
	if passphrase == "" {
		block, err = gossh.MarshalPrivateKey(private, "xenon test key")
	} else {
		block, err = gossh.MarshalPrivateKeyWithPassphrase(private, "xenon test key", []byte(passphrase))
	}
	if err != nil {
		return nil, err
	}
	return &SshKey{Signer: signer, Private: private, PrivatePem: pem.EncodeToMemory(block)}, nil
}

func (s *SshKey) AuthorizedKey() string {
	return string(gossh.MarshalAuthorizedKey(s.Signer.PublicKey()))
}
