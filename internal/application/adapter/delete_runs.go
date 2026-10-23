package adapter

import (
	"context"
)

type DeleteRunsService interface {
	Delete(ctx context.Context, localId, runId string) error
	Clear(ctx context.Context, localId, itemId string) error
}
