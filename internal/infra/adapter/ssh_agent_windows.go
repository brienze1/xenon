//go:build windows

package adapter

import (
	"io"
	"net"

	sshagent "github.com/xanzy/ssh-agent"
	"golang.org/x/crypto/ssh/agent"
)

type noConnection struct{}

func (noConnection) Close() error { return nil }

func connectSshAgent(socket string) (agent.Agent, io.Closer, error) {
	if socket != "" {
		return dialSshAgentSocket(socket)
	}
	client, connection, err := sshagent.New()
	if err != nil {
		return nil, nil, err
	}
	if connection == nil {
		return client, noConnection{}, nil
	}
	return client, connection, nil
}

func dialUnixSocket(socket string) (net.Conn, error) {
	return net.Dial("unix", socket)
}
