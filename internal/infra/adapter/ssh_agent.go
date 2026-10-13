package adapter

import (
	"io"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

type osSshAgent struct{}

func NewSshAgent() adapter.SshAgent {
	return &osSshAgent{}
}

func (o *osSshAgent) Signers() ([]ssh.Signer, io.Closer, error) {
	client, closer, err := connectSshAgent(properties.Properties().Git.SshAuthSock)
	if err != nil {
		return nil, nil, err
	}
	signers, err := client.Signers()
	if err != nil {
		_ = closer.Close()
		return nil, nil, err
	}
	return signers, closer, nil
}

func dialSshAgentSocket(socket string) (agent.Agent, io.Closer, error) {
	connection, err := dialUnixSocket(socket)
	if err != nil {
		return nil, nil, err
	}
	return agent.NewClient(connection), connection, nil
}
