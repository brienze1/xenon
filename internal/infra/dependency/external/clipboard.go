package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var clipboardOnce sync.Once
var clipboardInstance adapter.Clipboard

func GetClipboard() adapter.Clipboard {
	if clipboardInstance == nil {
		clipboardOnce.Do(func() {
			clipboardInstance = infraAdapter.NewClipboard()
		})
	}
	return clipboardInstance
}

func SetClipboard(clipboard adapter.Clipboard) {
	clipboardOnce.Do(func() {})
	clipboardInstance = clipboard
}
