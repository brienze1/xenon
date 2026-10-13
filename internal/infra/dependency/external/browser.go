package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var browserOnce sync.Once
var browserInstance adapter.Browser

func GetBrowser() adapter.Browser {
	if browserInstance == nil {
		browserOnce.Do(func() {
			browserInstance = infraAdapter.NewBrowser()
		})
	}
	return browserInstance
}

func SetBrowser(browser adapter.Browser) {
	browserOnce.Do(func() {})
	browserInstance = browser
}
