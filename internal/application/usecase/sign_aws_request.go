package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SignAwsRequest interface {
	Sign(ctx context.Context, request entity.EffectiveRequest, auth entity.AwsV4Auth) (*entity.EffectiveRequest, error)
}
