package usecase

import (
	"fmt"
	"strings"

	"github.com/brienze1/xenon/pkg/errs"
)

func RejectUnavailableAlias(aliasName string, environmentLabels []string) error {
	return errs.ForbiddenError(fmt.Sprintf("the alias %q is only for %s", aliasName, strings.Join(environmentLabels, ", ")), "ALS-09403")
}
