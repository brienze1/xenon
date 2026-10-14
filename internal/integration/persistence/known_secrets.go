package persistence

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type knownSecretsPersistence struct {
	known entity.KnownSecrets
	mu    sync.Mutex
}

func NewKnownSecretsPersistence() adapter.KnownSecretsPersistence {
	return &knownSecretsPersistence{known: entity.NewKnownSecrets()}
}

func (k *knownSecretsPersistence) Register(ctx context.Context, values ...string) entity.KnownSecrets {
	logger.Debug(ctx, "Started", map[string]any{"values": len(values)})

	k.mu.Lock()
	k.known = k.known.With(values...)
	snapshot := k.known
	k.mu.Unlock()

	logger.Debug(ctx, "Finished", snapshot)
	return snapshot
}
