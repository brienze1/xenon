package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

func MaskKnownSecrets(text string, known entity.KnownSecrets, collected ...string) string {
	return known.With(collected...).Mask(text)
}
