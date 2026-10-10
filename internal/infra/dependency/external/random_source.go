package external

import (
	"crypto/rand"
	mathrand "math/rand/v2"
	"sync"
)

var randomSourceOnce sync.Once
var randomSourceInstance mathrand.Source

func GetRandomSource() mathrand.Source {
	if randomSourceInstance == nil {
		randomSourceOnce.Do(func() {
			var seed [32]byte
			_, _ = rand.Read(seed[:])
			randomSourceInstance = mathrand.NewChaCha8(seed)
		})
	}
	return randomSourceInstance
}

func SetRandomSource(source mathrand.Source) {
	randomSourceOnce.Do(func() {})
	randomSourceInstance = source
}
