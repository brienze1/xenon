package gitaccountkind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	GitHub  enums.GitAccountKind = "github"
	Token   enums.GitAccountKind = "token"
	Ssh     enums.GitAccountKind = "ssh"
	Unknown enums.GitAccountKind = "unknown"
)

func GetGitAccountKind(s string) enums.GitAccountKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "github":
		return GitHub
	case "token":
		return Token
	case "ssh":
		return Ssh
	default:
		return Unknown
	}
}
