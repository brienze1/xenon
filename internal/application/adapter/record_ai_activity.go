package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type RecordAIActivityService interface {
	Record(ctx context.Context, localId, client string, activity entity.AIActivity)
}
