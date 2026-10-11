package nodekind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Mapping  enums.NodeKind = "mapping"
	Sequence enums.NodeKind = "sequence"
	Scalar   enums.NodeKind = "scalar"
	Unknown  enums.NodeKind = "unknown"
)

func GetNodeKind(s string) enums.NodeKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "mapping":
		return Mapping
	case "sequence":
		return Sequence
	case "scalar":
		return Scalar
	default:
		return Unknown
	}
}
