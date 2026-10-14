package errors

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func RejectDeniedAISend() error {
	return errs.ForbiddenError("The send was denied by the user in Xenon", entity.AISendDeniedCode)
}
