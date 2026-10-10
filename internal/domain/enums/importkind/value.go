package importkind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	PostmanCollection  enums.ImportKind = "postman_collection"
	PostmanEnvironment enums.ImportKind = "postman_environment"
	PostmanGlobals     enums.ImportKind = "postman_globals"
	PostmanDump        enums.ImportKind = "postman_dump"
	BrunoBru           enums.ImportKind = "bruno_bru"
	BrunoYaml          enums.ImportKind = "bruno_yaml"
	Curl               enums.ImportKind = "curl"
	Unknown            enums.ImportKind = "unknown"
)

func GetImportKind(s string) enums.ImportKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "postman_collection":
		return PostmanCollection
	case "postman_environment":
		return PostmanEnvironment
	case "postman_globals":
		return PostmanGlobals
	case "postman_dump":
		return PostmanDump
	case "bruno_bru":
		return BrunoBru
	case "bruno_yaml":
		return BrunoYaml
	case "curl":
		return Curl
	default:
		return Unknown
	}
}
