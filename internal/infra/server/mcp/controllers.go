package mcp

import "github.com/brienze1/xenon/internal/integration/adapter"

type Controllers struct {
	Call          adapter.MCPCallController
	Workspace     adapter.MCPToolController
	WorkspaceEdit adapter.MCPToolController
	Tree          adapter.MCPToolController
	Env           adapter.MCPToolController
	EnvEdit       adapter.MCPToolController
	TreeEdit      adapter.MCPToolController
	History       adapter.MCPToolController
	Docs          adapter.MCPToolController
	DocsEdit      adapter.MCPToolController
	Response      adapter.MCPToolController
	Snippet       adapter.MCPToolController
	Send          adapter.MCPToolController
	Cookies       adapter.MCPToolController
	CookiesEdit   adapter.MCPToolController
	Console       adapter.MCPToolController
	Git           adapter.MCPToolController
	GitEdit       adapter.MCPToolController
	GitPush       adapter.MCPToolController
	Scripts       adapter.MCPToolController
	Settings      adapter.MCPToolController
	SettingsEdit  adapter.MCPToolController
	Import        adapter.MCPToolController
	Export        adapter.MCPToolController
	Delete        adapter.MCPToolController
	Run           adapter.MCPToolController
	Runs          adapter.MCPToolController
}
