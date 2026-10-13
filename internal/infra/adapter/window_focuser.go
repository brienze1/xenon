package adapter

import (
	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

type wailsWindowFocuser struct{}

func NewWindowFocuser() adapter.WindowFocuser {
	return &wailsWindowFocuser{}
}

func (w *wailsWindowFocuser) Focus() {
	app := wails.Get()
	if app == nil {
		return
	}
	for _, window := range app.Window.GetAll() {
		window.Restore()
		window.Focus()
	}
}
