//go:build !windows

package adapter

import (
	"errors"
	"io"
	"net"

	"golang.org/x/crypto/ssh/agent"
)

var errSshAgentNotConfigured = errors.New("no SSH agent: SSH_AUTH_SOCK is not set")

func connectSshAgent(socket string) (agent.Agent, io.Closer, error) {
	if socket == "" {
		return nil, nil, errSshAgentNotConfigured
	}
	return dialSshAgentSocket(socket)
}

func dialUnixSocket(socket string) (net.Conn, error) {
	return net.Dial("unix", socket)
}
