package errors

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func RejectInvalidArguments(tool, reason string) error {
	return errs.BadRequestError("Invalid arguments for "+tool+": "+reason, entity.AIInvalidArgumentsCode)
}
