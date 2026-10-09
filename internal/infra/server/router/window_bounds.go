package router

import (
	"context"
	"sync"
	"time"

	wails "github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

const windowSaveDelay = 500 * time.Millisecond

func (r *router) restoreWindowBounds(app *wails.App, window *wails.WebviewWindow) {
	defer window.Show()
	state, err := r.window.Restore(context.Background(), dto.NewRestoreWindow(screensOf(app.Screen.GetAll())))
	if err != nil {
		return
	}
	window.SetBounds(wails.Rect{X: state.X, Y: state.Y, Width: state.Width, Height: state.Height})
	if state.Maximized {
		window.Maximise()
	}
}

func (r *router) rememberWindowBounds(window *wails.WebviewWindow) {
	var mu sync.Mutex
	var timer *time.Timer
	save := func() {
		bounds := window.Bounds()
		_ = r.window.Save(context.Background(), dto.NewWindowStateFromBounds(
			bounds.X, bounds.Y, bounds.Width, bounds.Height, window.IsMaximised(), window.IsMinimised(), window.IsFullscreen(),
		))
	}
	schedule := func(*wails.WindowEvent) {
		mu.Lock()
		defer mu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(windowSaveDelay, save)
	}
	for _, event := range []events.WindowEventType{events.Common.WindowDidMove, events.Common.WindowDidResize, events.Common.WindowMaximise, events.Common.WindowRestore} {
		window.OnWindowEvent(event, schedule)
	}
	window.RegisterHook(events.Common.WindowClosing, func(*wails.WindowEvent) {
		mu.Lock()
		if timer != nil {
			timer.Stop()
		}
		mu.Unlock()
		save()
	})
}

func screensOf(screens []*wails.Screen) []dto.Screen {
	result := make([]dto.Screen, 0, len(screens))
	for _, screen := range screens {
		area := screen.WorkArea
		result = append(result, dto.NewScreen(area.X, area.Y, area.Width, area.Height, screen.IsPrimary))
	}
	return result
}
