package binding

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/dependency/server"
	"github.com/brienze1/xenon/internal/infra/server/binding"
	"github.com/brienze1/xenon/internal/infra/server/mcp"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

var aiAccessServiceOnce sync.Once
var aiAccessServiceInstance *binding.AIAccessService

func GetAIAccessService() *binding.AIAccessService {
	if aiAccessServiceInstance == nil {
		aiAccessServiceOnce.Do(func() {
			server.GetMCPServer().Mount(mcpControllers())
			server.GetMCPServer().OnSessionClosed(denySessionApprovals)
			aiAccessServiceInstance = binding.NewAIAccessService(controller.GetAIAccessController())
		})
	}
	return aiAccessServiceInstance
}

func denySessionApprovals(ctx context.Context, sessionId string) {
	_ = controller.GetAIAccessController().DenyApprovals(ctx, dto.AIApprovalDenial{SessionId: sessionId})
}

func mcpControllers() mcp.Controllers {
	return mcp.Controllers{
		AIAccess:             controller.GetAIAccessController(),
		WorkspaceList:        controller.GetWorkspaceListController(),
		WorkspaceCreate:      controller.GetWorkspaceCreateController(),
		WorkspaceOpen:        controller.GetWorkspaceOpenController(),
		WorkspaceActivate:    controller.GetWorkspaceActivateController(),
		WorkspaceRename:      controller.GetWorkspaceRenameController(),
		WorkspaceRemove:      controller.GetWorkspaceRemoveController(),
		TreeSearch:           controller.GetTreeSearchController(),
		TreeManage:           controller.GetTreeManageController(),
		TreeSettings:         controller.GetTreeSettingsController(),
		TreeDocumentation:    controller.GetTreeDocumentationController(),
		RequestGet:           controller.GetRequestGetController(),
		RequestSave:          controller.GetRequestSaveController(),
		RequestSend:          controller.GetRequestSendController(),
		RequestResponse:      controller.GetRequestResponseController(),
		ExampleManage:        controller.GetExampleManageController(),
		OAuth:                controller.GetOAuthController(),
		EnvironmentList:      controller.GetEnvironmentListController(),
		EnvironmentManage:    controller.GetEnvironmentManageController(),
		EnvironmentSetActive: controller.GetEnvironmentSetActiveController(),
		VariableTable:        controller.GetVariableTableController(),
		VariableInspect:      controller.GetVariableInspectController(),
		History:              controller.GetHistoryController(),
		Snippet:              controller.GetSnippetController(),
		Cookie:               controller.GetCookieController(),
		Console:              controller.GetConsoleController(),
		SyncManage:           controller.GetSyncManageController(),
		SyncLink:             controller.GetSyncLinkController(),
		SyncTransfer:         controller.GetSyncTransferController(),
		SyncConflict:         controller.GetSyncConflictController(),
		ScriptTools:          controller.GetScriptToolsController(),
		Settings:             controller.GetSettingsController(),
		Import:               controller.GetImportManageController(),
		Export:               controller.GetExportController(),
	}
}
