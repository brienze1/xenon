package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindUpdateSession interface {
	FindUpdateSession(ctx context.Context) entity.UpdateSession
}
