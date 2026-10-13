package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	missingSecretsAttribute = "missing_secrets"
	lockedSecretsAttribute  = "locked_secrets"
)

func RejectUnavailableSecrets(preflight entity.SendPreflight) error {
	details := make([]errs.ErrorDetails, 0, 2)
	if preflight.HasMissingSecrets() {
		details = append(details, errs.ErrorDetails{Attribute: missingSecretsAttribute, Messages: preflight.MissingSecrets})
	}
	if preflight.HasLockedSecrets() {
		details = append(details, errs.ErrorDetails{Attribute: lockedSecretsAttribute, Messages: preflight.LockedSecrets})
	}
	return errs.ConflictError("the request uses secrets that have no value on this machine or are locked", "SND-06409", details...)
}
