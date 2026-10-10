package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
)

func MergeImportVariables(existing, imported []entity.VariableRow, resolution enums.ClashResolution) ([]entity.VariableRow, []entity.VariableRow) {
	merged := append([]entity.VariableRow{}, existing...)
	kept := make([]entity.VariableRow, 0)
	for _, row := range imported {
		index, clashes := variableIndex(merged, row.Key)
		switch {
		case !clashes:
			merged = append(merged, row)
		case resolution == clashresolution.Replace:
			merged[index] = row
		case resolution == clashresolution.KeepBoth:
			kept = append(kept, row)
		}
	}
	return merged, kept
}
