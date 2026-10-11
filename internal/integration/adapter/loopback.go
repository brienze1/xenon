package adapter

import "net"

type Loopback interface {
	Listen(address string) (net.Listener, error)
}
