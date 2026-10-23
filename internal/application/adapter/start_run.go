package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type StartRunService interface {
	Start(ctx context.Context, start entity.RunStart, appVersion string) (*entity.Run, error)
	// StartForAI resolves the target by path or item id, lays the overrides over its remembered configuration and asks the
	// user to approve an environment confirmation; the target comes back with every outcome, errors included.
	StartForAI(ctx context.Context, localId, path, itemId string, overrides entity.RunOverrides, appVersion, client, sessionId string) (entity.AIRunStart, error)
}
