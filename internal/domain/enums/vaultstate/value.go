package vaultstate

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Empty    enums.VaultState = "empty"
	Unlocked enums.VaultState = "unlocked"
	Locked   enums.VaultState = "locked"
	Broken   enums.VaultState = "broken"
	Unknown  enums.VaultState = "unknown"
)

func GetVaultState(s string) enums.VaultState {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "empty":
		return Empty
	case "unlocked":
		return Unlocked
	case "locked":
		return Locked
	case "broken":
		return Broken
	default:
		return Unknown
	}
}
