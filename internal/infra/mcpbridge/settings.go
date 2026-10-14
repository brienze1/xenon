package mcpbridge

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

func findAIAccess(ctx context.Context) entity.AIAccessSettings {
	settings, err := integration.GetRegistryPersistence().FindAppSettings(ctx)
	if err != nil {
		return entity.DefaultAIAccessSettings()
	}
	return settings.AIAccess
}
