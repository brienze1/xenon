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
	nextStepAttribute   = "next_step"
	offlineNextStep     = "Offline - check the connection and try again"
	restrictedNextStep  = "Ask an owner of the organisation to approve Xenon, or authorise single sign-on (SSO) for the token, or use a token account"
	protectedNextStep   = "Xenon does not run Git hooks or sign commits - allow Xenon's pushes on this branch, or sync another branch"
	rejectedNextStep    = "Check the host's message, fix the cause on the host, then push again"
	nonFastForwardStep  = "Pull, then push again"
	pushRejectionPrefix = "command error on"
	busyDescription     = "another sync operation is running on this workspace; wait for it to finish"
)

var (
	syncRemoteCodes    = remoteCodes{unauthorized: "GIT-11401", restricted: "GIT-12403", notFound: "GIT-15404", offline: "GIT-10503", failed: "GIT-26500"}
	restrictionMarkers = []string{"oauth app access restrictions", "saml", "single sign-on", "sso"}
	protectionMarkers  = []string{"protected branch", "protected ref", "signature", "signed", "hook declined", "gh006"}
)

type remoteCodes struct {
	unauthorized string
	restricted   string
	notFound     string
	offline      string
	failed       string
}

func remoteError(err error, codes remoteCodes) error {
	switch {
	case isOrganisationRestriction(err):
		return remoteFailure("the organisation restricts this account's access: "+err.Error(), codes.restricted, http.StatusForbidden, restrictedNextStep)
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

func pushError(err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, "non-fast-forward"):
		return errs.ConflictError("the remote branch moved since the last pull", "GIT-21409", errs.ErrorDetails{Attribute: nextStepAttribute, Messages: []string{nonFastForwardStep}})
	case strings.HasPrefix(message, pushRejectionPrefix) && mentionsAny(message, protectionMarkers):
		return remoteFailure("the host rejected the push: "+message, "GIT-14403", http.StatusForbidden, protectedNextStep)
	case strings.HasPrefix(message, pushRejectionPrefix):
		return remoteFailure("the host rejected the push: "+message, "GIT-13409", http.StatusConflict, rejectedNextStep)
	default:
		return remoteError(err, syncRemoteCodes)
	}
}

func isOrganisationRestriction(err error) bool {
	return errors.Is(err, transport.ErrAuthorizationFailed) && mentionsAny(err.Error(), restrictionMarkers)
}

func mentionsAny(message string, markers []string) bool {
	lowered := strings.ToLower(message)
	for _, marker := range markers {
		if strings.Contains(lowered, marker) {
			return true
		}
	}
	return false
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
