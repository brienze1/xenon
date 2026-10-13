package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/clashkind"
)

func FindEnvironmentClashes(itemKey, name string, existing []entity.Environment) []entity.Clash {
	match, found := entity.NamedEnvironment(existing, name)
	if !found {
		return []entity.Clash{}
	}
	return []entity.Clash{{ItemKey: itemKey, Kind: clashkind.Environment, ExistingID: match.ID, ExistingName: match.Name, Name: name}}
}
