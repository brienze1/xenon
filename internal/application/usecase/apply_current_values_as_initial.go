package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

func ApplyCurrentValuesAsInitial(rows []entity.VariableRow, current map[string]string) []entity.VariableRow {
	applied := make([]entity.VariableRow, 0, len(rows))
	for _, row := range rows {
		applied = append(applied, row.WithCurrentAsInitial(current))
	}
	return applied
}
