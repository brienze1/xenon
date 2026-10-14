package adapter

type AIAccessListener interface {
	Listen(port int) error
	Close()
	Clients() int
}
