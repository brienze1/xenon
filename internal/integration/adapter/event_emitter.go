package adapter

type EventEmitter interface {
	Emit(name string, data any)
}
