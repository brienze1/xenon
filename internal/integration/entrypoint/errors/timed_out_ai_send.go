package errors

import (
	"net/http"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func RejectTimedOutAISend() error {
	return errs.ClientError("The send was not made: approval timed out", entity.AISendTimedOutCode, http.StatusRequestTimeout)
}
