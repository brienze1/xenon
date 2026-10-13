package adapter

import (
	"net"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

type tcpLoopback struct{}

func NewLoopback() adapter.Loopback {
	return &tcpLoopback{}
}

func (t *tcpLoopback) Listen(address string) (net.Listener, error) {
	return net.Listen("tcp", address)
}
