package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type NotifyAIChangeService interface {
	Notify(ctx context.Context, localId string, changes []entity.AIChange) error
}
