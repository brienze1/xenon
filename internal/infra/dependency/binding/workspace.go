package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var workspaceServiceOnce sync.Once
var workspaceServiceInstance *binding.WorkspaceService

func GetWorkspaceService() *binding.WorkspaceService {
	if workspaceServiceInstance == nil {
		workspaceServiceOnce.Do(func() {
			workspaceServiceInstance = binding.NewWorkspaceService(
				controller.GetWorkspaceListController(),
				controller.GetWorkspaceCreateController(),
				controller.GetWorkspaceActivateController(),
				controller.GetWorkspaceOpenController(),
				controller.GetWorkspaceRenameController(),
				controller.GetWorkspaceRemoveController(),
				controller.GetWorkspaceLocateController(),
			)
		})
	}
	return workspaceServiceInstance
}
