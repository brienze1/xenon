package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var windowFocuserOnce sync.Once
var windowFocuserInstance adapter.WindowFocuser

func GetWindowFocuser() adapter.WindowFocuser {
	if windowFocuserInstance == nil {
		windowFocuserOnce.Do(func() {
			windowFocuserInstance = infraAdapter.NewWindowFocuser()
		})
	}
	return windowFocuserInstance
}

func SetWindowFocuser(windowFocuser adapter.WindowFocuser) {
	windowFocuserOnce.Do(func() {})
	windowFocuserInstance = windowFocuser
}
