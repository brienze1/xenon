package adapter

import (
	"io"

	"golang.org/x/crypto/ssh"
)

type SshAgent interface {
	Signers() ([]ssh.Signer, io.Closer, error)
}
