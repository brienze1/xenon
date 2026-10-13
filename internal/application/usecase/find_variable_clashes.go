package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/clashkind"
)

func FindVariableClashes(itemKey string, imported, existing []entity.VariableRow, ownerId string) []entity.Clash {
	clashes := make([]entity.Clash, 0)
	for _, row := range imported {
		if _, found := entity.VariableRowIndex(existing, row.Key); found {
			clashes = append(clashes, entity.Clash{ItemKey: itemKey, Kind: clashkind.Variable, ExistingID: ownerId, ExistingName: row.Key, Name: row.Key})
		}
	}
	return clashes
}
