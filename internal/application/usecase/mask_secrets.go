package usecase

import (
	"cmp"
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func MaskSecrets(text string, secrets []string) (string, bool) {
	ordered := slices.SortedStableFunc(slices.Values(secrets), func(a, b string) int {
		return cmp.Compare(len(b), len(a))
	})
	masked := text
	for _, secret := range ordered {
		if secret != "" {
			masked = strings.ReplaceAll(masked, secret, entity.MaskedValue)
		}
	}
	return masked, masked != text
}
