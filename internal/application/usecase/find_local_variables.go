package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindLocalVariables interface {
	Find(ctx context.Context, localId string) (*entity.LocalVariables, error)
}
