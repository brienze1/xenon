package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type scriptRequests struct {
	sendHttpRequest      usecase.SendHttpRequest
	generateDynamicValue usecase.GenerateDynamicValue
	variables            entity.VariableContext
	root                 string
	appVersion           string
	secrets              []string
}

func (s scriptRequests) Send(ctx context.Context, request entity.Request, localVariables map[string]any) entity.NestedRequest {
	logger.Info(ctx, "Started", map[string]any{"request": request})

	variables := s.variables.WithLocal(localVariables)
	preflight := usecase.CheckSendPreflight(request, variables, nil)
	if preflight.HasUnavailableSecrets() {
		rejected := entity.NewRejectedNestedRequest(preflight.UnavailableSecretsMessage())
		logger.Info(ctx, "Finished", map[string]any{"request": request, "result": rejected})
		return rejected
	}
	resolved, _ := usecase.ResolveRequestVariables(request, nil, variables, func(name string) (string, bool) {
		return s.generateDynamicValue.Generate(ctx, name)
	})
	effective := usecase.ResolveEffectiveRequest(resolved, nil, s.root)
	effective = effective.WithAutoHeaders(usecase.ComputeAutoHeaders(effective, s.appVersion))
	sent := entity.NewSentNestedRequest(effective, s.sendHttpRequest.Send(ctx, effective).WithMaskedValues(s.secrets))

	logger.Info(ctx, "Finished", map[string]any{"request": request, "result": sent})
	return sent
}
