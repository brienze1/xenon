package router

import (
	wails "github.com/wailsapp/wails/v3/pkg/application"
)

func (r *router) registerCookieRoutes() wails.Service {
	return wails.NewService(r.cookie)
}
