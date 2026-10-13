package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var loopbackOnce sync.Once
var loopbackInstance adapter.Loopback

func GetLoopback() adapter.Loopback {
	if loopbackInstance == nil {
		loopbackOnce.Do(func() {
			loopbackInstance = infraAdapter.NewLoopback()
		})
	}
	return loopbackInstance
}

func SetLoopback(loopback adapter.Loopback) {
	loopbackOnce.Do(func() {})
	loopbackInstance = loopback
}
