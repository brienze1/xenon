package usecase

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func FormatBulkRows(rows []entity.KeyValue) string {
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.IsBlank() {
			continue
		}
		line := row.Key + ":" + row.Value
		if !row.IsEnabled() {
			line = disabledRowPrefix + line
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
