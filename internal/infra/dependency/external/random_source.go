package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var randomSourceOnce sync.Once
var randomSourceInstance adapter.RandomSource

func GetRandomSource() adapter.RandomSource {
	if randomSourceInstance == nil {
		randomSourceOnce.Do(func() {
			randomSourceInstance = infraAdapter.NewRandomSource()
		})
	}
	return randomSourceInstance
}

func SetRandomSource(source adapter.RandomSource) {
	randomSourceOnce.Do(func() {})
	randomSourceInstance = source
}
