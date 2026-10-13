package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type BulkRows struct {
	Rows []KeyValue `json:"rows"`
}

func NewBulkRows(rows []entity.KeyValue) BulkRows {
	return BulkRows{Rows: NewKeyValues(rows)}
}

func (b *BulkRows) ToEntity() []entity.KeyValue {
	return keyValuesToEntity(b.Rows)
}

func (b BulkRows) LogValue() any {
	return map[string]any{"rows": len(b.Rows)}
}
