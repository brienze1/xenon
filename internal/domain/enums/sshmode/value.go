package sshmode

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Agent     enums.SshMode = "agent"
	KeyFile   enums.SshMode = "key_file"
	Generated enums.SshMode = "generated"
	Unknown   enums.SshMode = "unknown"
)

func GetSshMode(s string) enums.SshMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "agent":
		return Agent
	case "key_file":
		return KeyFile
	case "generated":
		return Generated
	default:
		return Unknown
	}
}
