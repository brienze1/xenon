package vaultmode

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Keychain   enums.VaultMode = "keychain"
	Passphrase enums.VaultMode = "passphrase"
	None       enums.VaultMode = "none"
	Unknown    enums.VaultMode = "unknown"
)

func GetVaultMode(s string) enums.VaultMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "keychain":
		return Keychain
	case "passphrase":
		return Passphrase
	case "none":
		return None
	default:
		return Unknown
	}
}
