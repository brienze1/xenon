package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	missingSecretsAttribute = "missing_secrets"
	lockedSecretsAttribute  = "locked_secrets"
)

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

// UsesSecretRows reports whether any scope of the context has secret rows, so the vault must be opened.
func UsesSecretRows(environment *entity.Environment, ancestors []entity.Collection, globals []entity.VariableRow) bool {
	if environment != nil && entity.HasSecretRows(environment.Variables) {
		return true
	}
	if len(ancestors) > 0 && entity.HasSecretRows(ancestors[0].OwnVariables()) {
		return true
	}
	return entity.HasSecretRows(globals)
}

func UnavailableSecretDetails(preflight entity.SendPreflight) []errs.ErrorDetails {
	details := make([]errs.ErrorDetails, 0, 2)
	if preflight.HasMissingSecrets() {
		details = append(details, errs.ErrorDetails{Attribute: missingSecretsAttribute, Messages: preflight.MissingSecrets})
	}
	if preflight.HasLockedSecrets() {
		details = append(details, errs.ErrorDetails{Attribute: lockedSecretsAttribute, Messages: preflight.LockedSecrets})
	}
	return details
}
