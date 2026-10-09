package formdatatype

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Text enums.FormDataType = "text"
	File enums.FormDataType = "file"
)

func GetFormDataType(s string) enums.FormDataType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "file":
		return File
	default:
		return Text
	}
}
