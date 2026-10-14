package errors

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func RejectUnsupportedAIAction(tool, action string) error {
	return errs.BadRequestError("The action \""+action+"\" of "+tool+" is not available", entity.AIInvalidArgumentsCode)
}
