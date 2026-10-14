package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateGetOAuthToken(request *dto.OAuthTokenReference) error {
	return validator.CustomValidator().Struct("OAU-16400", request)
}

func validateGetNewOAuthToken(request *dto.OAuthTokenRequest) error {
	return validator.CustomValidator().Struct("OAU-14400", request)
}

func validateRefreshOAuthToken(request *dto.OAuthTokenRequest) error {
	return validator.CustomValidator().Struct("OAU-15400", request)
}

func validateUseOAuthToken(request *dto.OAuthTokenUse) error {
	return validator.CustomValidator().Struct("OAU-19400", request)
}

func validateClearOAuthToken(request *dto.OAuthTokenReference) error {
	return validator.CustomValidator().Struct("OAU-20400", request)
}
