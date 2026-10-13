package usecase

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const disabledRowPrefix = "//"

func ParseBulkRows(text string) []entity.KeyValue {
	rows := make([]entity.KeyValue, 0)
	for _, line := range strings.Split(text, "\n") {
		content, disabled := strings.CutPrefix(strings.TrimSpace(line), disabledRowPrefix)
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}
		key, value, _ := strings.Cut(content, ":")
		row := entity.NewKeyValue(strings.TrimSpace(key), strings.TrimSpace(value))
		row.Enabled = !disabled
		rows = append(rows, row)
	}
	return rows
}
