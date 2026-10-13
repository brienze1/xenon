package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SendHttpRequest interface {
	Send(ctx context.Context, request entity.EffectiveRequest) entity.Response
}
