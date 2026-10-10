package router

import (
	wails "github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"github.com/brienze1/xenon/internal/integration/publisher/dto"
)

func (r *router) registerDropRoutes(app *wails.App, window *wails.WebviewWindow) {
	window.OnWindowEvent(events.Common.WindowFilesDropped, func(event *wails.WindowEvent) {
		app.Event.Emit(dto.ImportFilesDroppedEvent, dto.NewImportFilesDropped(event.Context().DroppedFiles()))
	})
}
