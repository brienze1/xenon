package router

import (
	wails "github.com/wailsapp/wails/v3/pkg/application"
)

func (r *router) registerConsoleRoutes() wails.Service {
	return wails.NewService(r.console)
}
