package entity

// AIOutputHasSecrets reports whether the leak guard had to mask anything in raw.
func AIOutputHasSecrets(raw, guarded string) bool {
	return guarded != raw
}
