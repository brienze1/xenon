package router

import (
	wails "github.com/wailsapp/wails/v3/pkg/application"
)

func (r *router) registerExampleRoutes() wails.Service {
	return wails.NewService(r.example)
}
