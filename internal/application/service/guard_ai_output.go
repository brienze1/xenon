package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type guardAIOutput struct {
	findVault            usecase.FindVault
	registerKnownSecrets usecase.RegisterKnownSecrets
}

func NewGuardAIOutputService(findVault usecase.FindVault, registerKnownSecrets usecase.RegisterKnownSecrets) adapter.GuardAIOutputService {
	return &guardAIOutput{findVault: findVault, registerKnownSecrets: registerKnownSecrets}
}

func (g *guardAIOutput) Guard(ctx context.Context, localId, text string) (string, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "length": len(text)})

	secrets, err := g.collect(ctx, localId)
	known := g.registerKnownSecrets.Register(ctx, secrets...)
	masked := usecase.MaskKnownSecrets(text, known)
	if err != nil {
		return masked, err
	}

	logger.Info(ctx, "Finished", known, map[string]any{"localId": localId, "length": len(text), "masked": len(masked)})
	return masked, nil
}

func (g *guardAIOutput) collect(ctx context.Context, localId string) ([]string, error) {
	if !entity.HasActiveWorkspace(localId) {
		return []string{}, nil
	}
	vault, err := g.findVault.Find(ctx, localId)
	if err != nil {
		return []string{}, err
	}
	return usecase.CollectKnownSecrets(vault), nil
}
