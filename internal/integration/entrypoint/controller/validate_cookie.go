package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateListCookies(request *dto.CookieJarReference) error {
	return validator.CustomValidator().Struct("CKE-01400", request)
}

func validateUpsertCookie(request *dto.UpsertCookie) error {
	return validator.CustomValidator().Struct("CKE-02400", request)
}

func validateGetCookieAllowlist(request *dto.CookieJarReference) error {
	return validator.CustomValidator().Struct("CKE-06400", request)
}

func validateSaveCookieAllowlist(request *dto.SaveCookieAllowlist) error {
	return validator.CustomValidator().Struct("CKE-07400", request)
}

func validateDeleteCookie(request *dto.DeleteCookie) error {
	return validator.CustomValidator().Struct("CKE-03400", request)
}

func validateDeleteCookieDomain(request *dto.DeleteCookieDomain) error {
	return validator.CustomValidator().Struct("CKE-04400", request)
}

func validateClearCookies(request *dto.CookieJarReference) error {
	return validator.CustomValidator().Struct("CKE-05400", request)
}
