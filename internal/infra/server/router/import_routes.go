package router

import (
	wails "github.com/wailsapp/wails/v3/pkg/application"
)

func (r *router) registerImportRoutes() wails.Service {
	return wails.NewService(r.importer)
}
