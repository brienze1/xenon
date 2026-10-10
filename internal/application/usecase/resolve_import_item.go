package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func ResolveImportItem(source entity.ImportSource, key string) (*entity.ImportItem, error) {
	item, found := source.Item(key)
	if !found {
		return nil, errs.NotFoundError("the import source has no item "+key, "IMP-16404")
	}
	return &item, nil
}
