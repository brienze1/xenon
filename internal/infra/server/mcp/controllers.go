package mcp

import "github.com/brienze1/xenon/internal/integration/adapter"

type Controllers struct {
	AIAccess             adapter.AIAccessController
	WorkspaceList        adapter.WorkspaceListController
	WorkspaceCreate      adapter.WorkspaceCreateController
	WorkspaceOpen        adapter.WorkspaceOpenController
	WorkspaceActivate    adapter.WorkspaceActivateController
	WorkspaceRename      adapter.WorkspaceRenameController
	WorkspaceRemove      adapter.WorkspaceRemoveController
	TreeSearch           adapter.TreeSearchController
	TreeManage           adapter.TreeManageController
	TreeSettings         adapter.TreeSettingsController
	TreeDocumentation    adapter.TreeDocumentationController
	RequestGet           adapter.RequestGetController
	RequestSave          adapter.RequestSaveController
	RequestSend          adapter.RequestSendController
	RequestResponse      adapter.RequestResponseController
	ExampleManage        adapter.ExampleManageController
	OAuth                adapter.OAuthController
	EnvironmentList      adapter.EnvironmentListController
	EnvironmentManage    adapter.EnvironmentManageController
	EnvironmentSetActive adapter.EnvironmentSetActiveController
	VariableTable        adapter.VariableTableController
	VariableInspect      adapter.VariableInspectController
	History              adapter.HistoryController
	Snippet              adapter.SnippetController
	Cookie               adapter.CookieController
	Console              adapter.ConsoleController
	SyncManage           adapter.SyncManageController
	SyncLink             adapter.SyncLinkController
	SyncTransfer         adapter.SyncTransferController
	SyncConflict         adapter.SyncConflictController
	ScriptTools          adapter.ScriptToolsController
	Settings             adapter.SettingsController
	Import               adapter.ImportManageController
	Export               adapter.ExportController
}
