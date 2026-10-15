package importkind

import "github.com/brienze1/xenon/internal/domain/enums"

const (
	PostmanCollection    enums.ImportKind = "postman_collection"
	PostmanEnvironment   enums.ImportKind = "postman_environment"
	PostmanGlobals       enums.ImportKind = "postman_globals"
	PostmanDump          enums.ImportKind = "postman_dump"
	PostmanRequest       enums.ImportKind = "postman_request"
	PostmanWorkspace     enums.ImportKind = "postman_workspace"
	BrunoBru             enums.ImportKind = "bruno_bru"
	BrunoYaml            enums.ImportKind = "bruno_yaml"
	Curl                 enums.ImportKind = "curl"
	XenonWorkspaceZip    enums.ImportKind = "xenon_workspace_zip"
	XenonWorkspaceFolder enums.ImportKind = "xenon_workspace_folder"
	XenonBundle          enums.ImportKind = "xenon_bundle"
	XenonZip             enums.ImportKind = "xenon_zip"
	XenonRequest         enums.ImportKind = "xenon_request"
	XenonEnvironment     enums.ImportKind = "xenon_environment"
	Folder               enums.ImportKind = "folder"
	Mixed                enums.ImportKind = "mixed"
	Unknown              enums.ImportKind = "unknown"
)
