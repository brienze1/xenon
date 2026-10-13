package router

import (
	"runtime"

	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon/internal/integration/publisher/dto"
)

const menuSource = "menu"

func (r *router) registerMenuRoutes(app *wails.App) {
	if runtime.GOOS != "darwin" {
		return
	}
	menu := app.NewMenu()
	menu.AddRole(wails.AppMenu)
	file := menu.AddSubmenu("File")
	file.Add("Import…").SetAccelerator("CmdOrCtrl+O").OnClick(func(*wails.Context) {
		app.Event.Emit(dto.ImportRequestedEvent, dto.NewImportRequested(menuSource))
	})
	file.Add("Export Workspace…").OnClick(func(*wails.Context) {
		app.Event.Emit(dto.ExportRequestedEvent, dto.NewExportRequested(menuSource))
	})
	file.AddSeparator()
	file.AddRole(wails.CloseWindow)
	menu.AddRole(wails.EditMenu)
	// The UI zooms itself (the layout grows to fit the window); the built-in zoom roles magnify the page and add scrollbars.
	view := menu.AddSubmenu("View")
	view.Add("Actual Size").SetAccelerator("CmdOrCtrl+0").OnClick(func(*wails.Context) {
		app.Event.Emit(dto.ViewZoomRequestedEvent, dto.NewViewZoomRequested("reset"))
	})
	view.Add("Zoom In").SetAccelerator("CmdOrCtrl+plus").OnClick(func(*wails.Context) {
		app.Event.Emit(dto.ViewZoomRequestedEvent, dto.NewViewZoomRequested("in"))
	})
	view.Add("Zoom Out").SetAccelerator("CmdOrCtrl+-").OnClick(func(*wails.Context) {
		app.Event.Emit(dto.ViewZoomRequestedEvent, dto.NewViewZoomRequested("out"))
	})
	view.AddSeparator()
	view.AddRole(wails.ToggleFullscreen)
	menu.AddRole(wails.WindowMenu)
	menu.AddRole(wails.HelpMenu)
	app.Menu.Set(menu)
}
