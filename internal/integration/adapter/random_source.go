package adapter

type RandomSource interface {
	Uint64() uint64
}
