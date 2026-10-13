package usecase

import (
	"fmt"

	"github.com/brienze1/xenon/pkg/errs"
)

func RejectScriptCookieAccess(host string) error {
	return errs.ForbiddenError(fmt.Sprintf("CookieStore: programmatic access to %q is denied", host), "CKE-09403")
}
