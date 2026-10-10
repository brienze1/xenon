package importorigin

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Postman enums.ImportOrigin = "postman"
	Bruno   enums.ImportOrigin = "bruno"
	Unknown enums.ImportOrigin = "unknown"
)

func GetImportOrigin(s string) enums.ImportOrigin {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "postman":
		return Postman
	case "bruno":
		return Bruno
	default:
		return Unknown
	}
}
