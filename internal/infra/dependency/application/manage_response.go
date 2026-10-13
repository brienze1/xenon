package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageResponseOnce sync.Once
var manageResponseInstance adapter.ManageResponseService

func GetManageResponseService() adapter.ManageResponseService {
	if manageResponseInstance == nil {
		manageResponseOnce.Do(func() {
			manageResponseInstance = service.NewManageResponseService(
				integration.GetResponseBodyPersistence(),
				integration.GetResponseBodyPersistence(),
				integration.GetResponseBodyPersistence(),
			)
		})
	}
	return manageResponseInstance
}
