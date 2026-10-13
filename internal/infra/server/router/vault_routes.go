package router

import (
	wails "github.com/wailsapp/wails/v3/pkg/application"
)

func (r *router) registerVaultRoutes() wails.Service {
	return wails.NewService(r.vault)
}
