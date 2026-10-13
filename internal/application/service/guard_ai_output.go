package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/pkg/logger"
)

type guardAIOutput struct {
	findVault            usecase.FindVault
	registerKnownSecrets usecase.RegisterKnownSecrets
}

func NewGuardAIOutputService(findVault usecase.FindVault, registerKnownSecrets usecase.RegisterKnownSecrets) adapter.GuardAIOutputService {
	return &guardAIOutput{findVault: findVault, registerKnownSecrets: registerKnownSecrets}
}

func (g *guardAIOutput) Guard(ctx context.Context, localId, text string) string {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "length": len(text)})

	known := g.registerKnownSecrets.Register(ctx, g.collect(ctx, localId)...)
	masked := usecase.MaskKnownSecrets(text, known)

	logger.Info(ctx, "Finished", known, map[string]any{"localId": localId, "length": len(masked)})
	return masked
}

func (g *guardAIOutput) Register(ctx context.Context, values ...string) {
	logger.Info(ctx, "Started", map[string]any{"values": len(values)})

	known := g.registerKnownSecrets.Register(ctx, values...)

	logger.Info(ctx, "Finished", known)
}

func (g *guardAIOutput) collect(ctx context.Context, localId string) []string {
	if localId == "" {
		return []string{}
	}
	vault, err := g.findVault.Find(ctx, localId)
	if err != nil {
		logger.Warn(ctx, err, "The vault could not be read; masking with the secrets seen so far", map[string]any{"localId": localId})
		return []string{}
	}
	return usecase.CollectKnownSecrets(vault)
}
