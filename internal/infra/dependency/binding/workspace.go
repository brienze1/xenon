package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var workspaceBindingOnce sync.Once
var workspaceBindingInstance *binding.WorkspaceService

func GetWorkspaceBinding() *binding.WorkspaceService {
	if workspaceBindingInstance == nil {
		workspaceBindingOnce.Do(func() {
			workspaceBindingInstance = binding.NewWorkspaceService(
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
	return workspaceBindingInstance
}
