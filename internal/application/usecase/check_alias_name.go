package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

// CheckAliasName applies the alias name rules: 1 to 200 characters, not "Default", not another alias's name.
func CheckAliasName(request entity.Request, alias entity.RequestAlias) error {
	switch {
	case !alias.HasValidName():
		return RejectInvalidAliasName()
	case alias.HasReservedName():
		return RejectReservedAliasName()
	case request.HasAliasNamed(alias.Name, alias.ID):
		return RejectTakenAliasName()
	}
	return nil
}
