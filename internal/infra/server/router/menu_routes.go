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
	menu.AddRole(wails.ViewMenu)
	menu.AddRole(wails.WindowMenu)
	menu.AddRole(wails.HelpMenu)
	app.Menu.Set(menu)
}
