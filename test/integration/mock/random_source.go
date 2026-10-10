package mock

import (
	"math/rand/v2"
	"sync"
)

const randomSeed = 20261009

type RandomSource struct {
	source *rand.PCG
	mu     sync.Mutex
}

func NewRandomSource() *RandomSource {
	return &RandomSource{source: rand.NewPCG(randomSeed, randomSeed)}
}

func (r *RandomSource) Uint64() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.source.Uint64()
}

func (r *RandomSource) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.source.Seed(randomSeed, randomSeed)
}
