package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var searchTreeOnce sync.Once
var searchTreeInstance adapter.SearchTreeService

func GetSearchTreeService() adapter.SearchTreeService {
	if searchTreeInstance == nil {
		searchTreeOnce.Do(func() {
			searchTreeInstance = service.NewSearchTreeService(integration.GetRegistryPersistence(), integration.GetTreePersistence())
		})
	}
	return searchTreeInstance
}
