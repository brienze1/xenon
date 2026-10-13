package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveLocalVariables interface {
	Save(ctx context.Context, localId string, local entity.LocalVariables) error
}
