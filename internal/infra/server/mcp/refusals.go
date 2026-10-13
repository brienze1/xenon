package mcp

import (
	"errors"
	"net/http"

	"github.com/brienze1/xenon/pkg/errs"
)

func sendDenied() error {
	return errs.ForbiddenError("The send was denied by the user in Xenon", deniedCode)
}

func approvalTimedOut() error {
	return errs.ClientError("The send was not made: approval timed out", timedOutCode, http.StatusRequestTimeout)
}

func snippetSecretsRefused() error {
	return humanOnly("Generating a snippet with secrets")
}

func responseBodySecretRefused() error {
	return humanOnly("Saving a binary response body that contains a secret")
}

func failedWithSendCode(err error, code string) bool {
	var custom errs.BaseError
	return errors.As(err, &custom) && custom.Code() == code
}
