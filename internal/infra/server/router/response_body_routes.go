package router

import (
	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

func (r *router) registerResponseBodyRoutes() wails.Service {
	return wails.NewServiceWithOptions(r.responseBody, wails.ServiceOptions{Route: dto.ResponseBodyRoute})
}
