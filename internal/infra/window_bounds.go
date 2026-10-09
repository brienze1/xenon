package application

import (
	"context"
	"sync"
	"time"

	wails "github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"github.com/brienze1/xenon/internal/infra/server/binding"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

const windowSaveDelay = 500 * time.Millisecond

func restoreWindowBounds(app *wails.App, window *wails.WebviewWindow, windowService *binding.WindowService) {
	defer window.Show()
	state, err := windowService.Restore(context.Background(), dto.RestoreWindow{Screens: screensOf(app.Screen.GetAll())})
	if err != nil {
		return
	}
	window.SetBounds(wails.Rect{X: state.X, Y: state.Y, Width: state.Width, Height: state.Height})
	if state.Maximized {
		window.Maximise()
	}
}

func rememberWindowBounds(window *wails.WebviewWindow, windowService *binding.WindowService) {
	var mu sync.Mutex
	var timer *time.Timer
	save := func() {
		if window.IsMinimised() || window.IsFullscreen() {
			return
		}
		bounds := window.Bounds()
		_ = windowService.Save(context.Background(), dto.WindowState{
			X:         bounds.X,
			Y:         bounds.Y,
			Width:     bounds.Width,
			Height:    bounds.Height,
			Maximized: window.IsMaximised(),
		})
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
		result = append(result, dto.Screen{X: area.X, Y: area.Y, Width: area.Width, Height: area.Height, Primary: screen.IsPrimary})
	}
	return result
}
