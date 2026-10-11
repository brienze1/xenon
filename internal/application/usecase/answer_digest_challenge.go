package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type AnswerDigestChallenge interface {
	ReadChallenge(ctx context.Context, challenge string) (*entity.DigestState, error)
	Answer(ctx context.Context, request entity.EffectiveRequest, auth entity.DigestAuth, state entity.DigestState) (string, error)
}
