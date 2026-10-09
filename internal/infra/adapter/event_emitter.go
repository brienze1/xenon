package adapter

import (
	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

type wailsEventEmitter struct{}

func NewEventEmitter() adapter.EventEmitter {
	return &wailsEventEmitter{}
}

func (w *wailsEventEmitter) Emit(name string, data any) {
	app := wails.Get()
	if app == nil {
		return
	}
	app.Event.Emit(name, data)
}
