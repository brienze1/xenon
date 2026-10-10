package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

// CheckSendPreflight lists the secrets the request uses (inherited auth and headers included, nested values followed) that have no value or sit in a locked vault.
func CheckSendPreflight(request entity.Request, ancestors []entity.Collection, context entity.VariableContext) entity.SendPreflight {
	preflight := entity.SendPreflight{}
	inherited := request
	inherited.Auth = effectiveAuth(request, ancestors)
	inherited.Headers = effectiveHeaders(request, ancestors)
	inherited.MapText(func(text string) string {
		preflight = unavailableSecrets(text, context, preflight, maxResolutionPasses)
		return text
	})
	return preflight
}

func unavailableSecrets(text string, context entity.VariableContext, preflight entity.SendPreflight, passesLeft int) entity.SendPreflight {
	if passesLeft == 0 {
		return preflight
	}
	for _, reference := range TokenizeReferences("", text) {
		if variable, found := context.Lookup(reference.Name); found && variable.IsUnavailableSecret() {
			preflight = preflight.WithUnavailable(variable)
		}
	}
	next := substituteVariables(text, context, nil, 1)
	if next == text {
		return preflight
	}
	return unavailableSecrets(next, context, preflight, passesLeft-1)
}
