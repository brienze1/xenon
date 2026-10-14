package errors

import (
	"errors"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	environmentIdAttribute = "environment_id"
	environmentAttribute   = "name"
	colorAttribute         = "color"
)

var unconfirmedSend = usecase.RejectUnconfirmedSend(entity.SendPreflight{ConfirmEnvironment: &entity.Environment{}}).(errs.BaseError)

// UnconfirmedSendEnvironment reads the environment out of a usecase.RejectUnconfirmedSend error; false for any other error.
func UnconfirmedSendEnvironment(err error) (entity.AIApprovalEnvironment, bool) {
	var custom errs.BaseError
	if !errors.As(err, &custom) || custom.Code() != unconfirmedSend.Code() {
		return entity.AIApprovalEnvironment{}, false
	}
	details := map[string]string{}
	for _, detail := range custom.ToDto("").Error.ErrorDetails {
		if len(detail.Messages) > 0 {
			details[detail.Attribute] = detail.Messages[0]
		}
	}
	return entity.NewAIApprovalEnvironment(details[environmentIdAttribute], details[environmentAttribute], details[colorAttribute]), true
}
