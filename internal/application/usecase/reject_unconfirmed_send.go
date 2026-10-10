package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	environmentIdAttribute = "environment_id"
	environmentAttribute   = "name"
	colorAttribute         = "color"
)

func RejectUnconfirmedSend(preflight entity.SendPreflight) error {
	environment := preflight.ConfirmEnvironment
	return errs.ConflictError("the active environment asks for a confirmation before every send", "SND-05409",
		errs.ErrorDetails{Attribute: environmentIdAttribute, Messages: []string{environment.ID}},
		errs.ErrorDetails{Attribute: environmentAttribute, Messages: []string{environment.Name}},
		errs.ErrorDetails{Attribute: colorAttribute, Messages: []string{environment.Color.String()}},
	)
}
