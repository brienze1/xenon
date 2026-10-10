package adapter

import (
	"crypto/rand"
	mathrand "math/rand/v2"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

func NewRandomSource() adapter.RandomSource {
	var seed [32]byte
	_, _ = rand.Read(seed[:])
	return mathrand.NewChaCha8(seed)
}
