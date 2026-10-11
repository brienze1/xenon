package usecase

import (
	"net/http"

	"github.com/brienze1/xenon/pkg/errs"
)

const (
	providerErrorAttribute       = "error"
	providerDescriptionAttribute = "error_description"
	ownerAttribute               = "owner_id"
)

func RejectInvalidOAuthConfig(reason string) error {
	return errs.BadRequestError("the OAuth 2.0 configuration cannot obtain a token: "+reason, "OAU-01400")
}

func RejectOAuthProviderError(code, description string) error {
	err := errs.BadRequestError("the OAuth 2.0 provider refused the request: "+code, "OAU-04401",
		errs.ErrorDetails{Attribute: providerErrorAttribute, Messages: []string{code}},
		errs.ErrorDetails{Attribute: providerDescriptionAttribute, Messages: []string{description}},
	)
	err.SetStatusCode(http.StatusUnauthorized, err.Description())
	return err
}

func RejectOAuthStateMismatch() error {
	return errs.BadRequestError("the OAuth 2.0 redirect carried another state; the token was not requested", "OAU-05400")
}

func RejectMissingRefreshToken() error {
	return errs.NotFoundError("there is no refresh token for this auth; get a new access token", "OAU-06404")
}

func RejectMissingOAuthToken() error {
	return errs.NotFoundError("there is no OAuth 2.0 token for this auth", "OAU-10404")
}

func RejectExpiredToken(owner string) error {
	return errs.ConflictError("the OAuth 2.0 access token expired and cannot be refreshed; get a new access token or send anyway", "SND-07409",
		errs.ErrorDetails{Attribute: ownerAttribute, Messages: []string{owner}},
	)
}
