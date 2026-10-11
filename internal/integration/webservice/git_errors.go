package webservice

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/brienze1/xenon/pkg/errs"
)

const (
	nextStepAttribute = "next_step"
	offlineNextStep   = "Offline - check the connection and try again"
)

type remoteCodes struct {
	unauthorized string
	notFound     string
	offline      string
	failed       string
}

func remoteError(err error, codes remoteCodes) error {
	switch {
	case isAuthenticationFailure(err):
		return remoteFailure("the Git host refused the credentials", codes.unauthorized, http.StatusUnauthorized,
			"Sign in again, or replace the token or SSH key of this account")
	case errors.Is(err, transport.ErrRepositoryNotFound):
		return remoteFailure("the repository was not found or this account cannot see it", codes.notFound, http.StatusNotFound,
			"Check the repository URL, or choose an account that has access to it")
	case isOffline(err):
		return remoteFailure("the Git host could not be reached: "+err.Error(), codes.offline, http.StatusServiceUnavailable,
			offlineNextStep)
	default:
		return errs.InternalServerError(err, codes.failed, "the Git operation failed: "+err.Error())
	}
}

func remoteFailure(description, code string, status int, nextStep string) error {
	err := errs.BadRequestError(description, code, errs.ErrorDetails{Attribute: nextStepAttribute, Messages: []string{nextStep}})
	err.SetStatusCode(status, description)
	return err
}

func isAuthenticationFailure(err error) bool {
	return errors.Is(err, transport.ErrAuthenticationRequired) || errors.Is(err, transport.ErrAuthorizationFailed) ||
		strings.Contains(err.Error(), "unable to authenticate")
}

func isOffline(err error) bool {
	var netError net.Error
	var opError *net.OpError
	return errors.Is(err, context.DeadlineExceeded) || errors.As(err, &opError) || (errors.As(err, &netError) && netError.Timeout())
}
