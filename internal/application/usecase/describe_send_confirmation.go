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

func DescribeSendConfirmation(preflight entity.SendPreflight) []errs.ErrorDetails {
	environment := preflight.ConfirmEnvironment
	return []errs.ErrorDetails{
		{Attribute: environmentIdAttribute, Messages: []string{environment.ID}},
		{Attribute: environmentAttribute, Messages: []string{environment.Name}},
		{Attribute: colorAttribute, Messages: []string{environment.Color.String()}},
	}
}
