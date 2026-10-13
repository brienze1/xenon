package mcp

import (
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	actionArgument = "action"
	targetArgument = "target"
)

type effect int

const (
	readOnly effect = iota
	write
	destructive
)

type tool struct {
	handler         func(*tools) Handler
	name            string
	description     string
	selector        string
	actions         []string
	effect          effect
	idempotent      bool
	openWorld       bool
	workspaceScoped bool
}

var registry = []tool{
	{name: "workspace", description: "List workspaces or show the active one.", effect: readOnly, idempotent: true, actions: []string{"list", "active"}, handler: func(t *tools) Handler { return t.workspace }},
	{name: "tree", description: "Browse, read and search collections, folders, requests and examples.", effect: readOnly, idempotent: true, workspaceScoped: true, actions: []string{"list", "get", "settings", "search", "token"}, handler: func(t *tools) Handler { return t.tree }},
	{name: "env", description: "List environments and read or inspect variables.", effect: readOnly, idempotent: true, workspaceScoped: true, actions: []string{"list", "get", "inspect"}, handler: func(t *tools) Handler { return t.env }},
	{name: "history", description: "List and read History (sends and AI activity).", effect: readOnly, idempotent: true, workspaceScoped: true, actions: []string{"list", "get"}, handler: func(t *tools) Handler { return t.history }},
	{name: "docs", description: "Read a collection's OpenAPI document or preview the plan for an edited one.", effect: readOnly, idempotent: true, workspaceScoped: true, actions: []string{"get", "plan"}, handler: func(t *tools) Handler { return t.docs }},
	{name: "response", description: "Read the last response of an AI send, including more of a truncated body.", effect: readOnly, idempotent: true, workspaceScoped: true, actions: []string{"last", "body"}, handler: func(t *tools) Handler { return t.response }},
	{name: "snippet", description: "Generate a code snippet for a request (secrets masked).", effect: readOnly, idempotent: true, workspaceScoped: true, actions: []string{"languages", "generate"}, handler: func(t *tools) Handler { return t.snippet }},
	{name: "cookies", description: "List cookies and the cookie allowlist.", effect: readOnly, idempotent: true, workspaceScoped: true, actions: []string{"list", "allowlist"}, handler: func(t *tools) Handler { return t.cookies }},
	{name: "console", description: "Read console entries from sends and scripts.", effect: readOnly, idempotent: true, workspaceScoped: true, actions: []string{"list", "get"}, handler: func(t *tools) Handler { return t.console }},
	{name: "git", description: "Show git sync status and pending changes.", effect: readOnly, idempotent: true, workspaceScoped: true, actions: []string{"status", "changes", "session"}, handler: func(t *tools) Handler { return t.git }},
	{name: "scripts", description: "Check a script for errors or read the scripting reference.", effect: readOnly, idempotent: true, workspaceScoped: true, actions: []string{"check", "reference", "snippets"}, handler: func(t *tools) Handler { return t.scripts }},
	{name: "settings", description: "Read app settings.", effect: readOnly, idempotent: true, workspaceScoped: true, actions: []string{"get"}, handler: func(t *tools) Handler { return t.settings }},
	{name: "workspace_edit", description: "Create, open, switch or rename a workspace.", effect: write, actions: []string{"create", "open", "switch", "rename"}, handler: func(t *tools) Handler { return t.workspaceEdit }},
	{name: "tree_edit", description: "Create, update, rename, move or duplicate items, edit collection/folder settings, and manage OAuth tokens.", effect: destructive, openWorld: true, workspaceScoped: true, actions: []string{"create", "update", "rename", "move", "duplicate", "save_settings", "get_token", "refresh_token", "use_token"}, handler: func(t *tools) Handler { return t.treeEdit }},
	{name: "env_edit", description: "Create or update environments, switch the active one, and edit variables (secret values are write-only).", effect: destructive, workspaceScoped: true, actions: []string{"create", "update", "duplicate", "set_active", "add_var", "set_vars", "set_current", "reset", "persist"}, handler: func(t *tools) Handler { return t.envEdit }},
	{name: "docs_edit", description: "Apply an edited OpenAPI document to a collection.", effect: destructive, workspaceScoped: true, actions: []string{"apply"}, handler: func(t *tools) Handler { return t.docsEdit }},
	{name: "cookies_edit", description: "Add or edit cookies and the cookie allowlist.", effect: destructive, idempotent: true, workspaceScoped: true, actions: []string{"upsert", "allowlist"}, handler: func(t *tools) Handler { return t.cookiesEdit }},
	{name: "git_edit", description: "Link, configure, commit, pull and resolve git sync for the workspace.", effect: destructive, openWorld: true, workspaceScoped: true, actions: []string{"link", "clone", "settings", "commit", "pull", "resolve", "finish_pull", "cancel_pull"}, handler: func(t *tools) Handler { return t.gitEdit }},
	{name: "settings_edit", description: "Change app settings (script timeout, history caps).", effect: write, idempotent: true, workspaceScoped: true, actions: []string{"save"}, handler: func(t *tools) Handler { return t.settingsEdit }},
	{name: "import", description: "Import Postman, Bruno or Xenon files from paths into the workspace.", effect: destructive, workspaceScoped: true, actions: []string{"detect", "clashes", "run", "discard"}, handler: func(t *tools) Handler { return t.importFiles }},
	{name: "export", description: "Export an item, environment or the workspace to a file path (never with secrets), or save a response body.", effect: destructive, workspaceScoped: true, actions: []string{"options", "run", "response_body"}, handler: func(t *tools) Handler { return t.export }},
	{name: "send", description: "Send a saved or ad-hoc request with the active environment and return the response.", effect: destructive, openWorld: true, workspaceScoped: true, handler: func(t *tools) Handler { return t.send }},
	{name: "delete", description: "Delete an item, environment, cookie(s), workspace, OAuth token, git link, History or the console.", effect: destructive, idempotent: true, workspaceScoped: true, selector: targetArgument, actions: []string{"item", "environment", "cookie", "cookie_domain", "cookies", "workspace", "token", "git_link", "history", "console"}, handler: func(t *tools) Handler { return t.delete }},
	{name: "git_push", description: "Push committed changes to the linked remote.", effect: destructive, openWorld: true, workspaceScoped: true, handler: func(t *tools) Handler { return t.gitPush }},
}

func (t tool) argumentName() string {
	if t.selector != "" {
		return t.selector
	}
	return actionArgument
}

func (t tool) hasActions() bool {
	return len(t.actions) > 0
}

func (t tool) defaultAction() string {
	if len(t.actions) == 1 {
		return t.actions[0]
	}
	return ""
}

func (t tool) allows(action string) bool {
	for _, candidate := range t.actions {
		if candidate == action {
			return true
		}
	}
	return false
}

func (t tool) definition() *sdk.Tool {
	return &sdk.Tool{
		Name:        t.name,
		Description: t.description,
		InputSchema: t.inputSchema(),
		Annotations: t.annotations(),
	}
}

func (t tool) inputSchema() map[string]any {
	schema := map[string]any{"type": "object"}
	if !t.hasActions() {
		return schema
	}
	schema["properties"] = map[string]any{t.argumentName(): map[string]any{"type": "string", "enum": t.actions}}
	if t.defaultAction() == "" {
		schema["required"] = []string{t.argumentName()}
	}
	return schema
}

func (t tool) annotations() *sdk.ToolAnnotations {
	destructiveHint := t.effect == destructive
	openWorldHint := t.openWorld
	annotations := &sdk.ToolAnnotations{
		ReadOnlyHint:   t.effect == readOnly,
		IdempotentHint: t.idempotent,
		OpenWorldHint:  &openWorldHint,
	}
	if t.effect != readOnly {
		annotations.DestructiveHint = &destructiveHint
	}
	return annotations
}
