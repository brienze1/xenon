package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func ResolveAlias(request entity.Request, aliasId string) (*entity.RequestAlias, error) {
	alias, found := request.Alias(aliasId)
	if !found {
		return nil, errs.NotFoundError("the alias does not exist", "ALS-01404")
	}
	return &alias, nil
}
