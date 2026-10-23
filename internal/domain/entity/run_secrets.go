package entity

import "slices"

// MergeSecrets appends the non-empty secrets not already known, keeping their order.
func MergeSecrets(known, secrets []string) []string {
	for _, secret := range secrets {
		if secret != "" && !slices.Contains(known, secret) {
			known = append(known, secret)
		}
	}
	return known
}
