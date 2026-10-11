package importkind

import "github.com/brienze1/xenon/internal/domain/enums"

const (
	PostmanCollection    enums.ImportKind = "postman_collection"
	PostmanEnvironment   enums.ImportKind = "postman_environment"
	PostmanGlobals       enums.ImportKind = "postman_globals"
	PostmanDump          enums.ImportKind = "postman_dump"
	BrunoBru             enums.ImportKind = "bruno_bru"
	BrunoYaml            enums.ImportKind = "bruno_yaml"
	Curl                 enums.ImportKind = "curl"
	XenonWorkspaceZip    enums.ImportKind = "xenon_workspace_zip"
	XenonWorkspaceFolder enums.ImportKind = "xenon_workspace_folder"
	XenonBundle          enums.ImportKind = "xenon_bundle"
	XenonZip             enums.ImportKind = "xenon_zip"
	XenonRequest         enums.ImportKind = "xenon_request"
	XenonEnvironment     enums.ImportKind = "xenon_environment"
	Unknown              enums.ImportKind = "unknown"
)
