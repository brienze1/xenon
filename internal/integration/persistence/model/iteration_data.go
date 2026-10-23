package model

import "github.com/brienze1/xenon/internal/domain/entity"

type IterationData struct {
	Rows []map[string]any
}

func NewIterationData(rows []map[string]any) *IterationData {
	return &IterationData{Rows: rows}
}

func (i *IterationData) ToEntity(fileName string) entity.IterationData {
	return entity.IterationData{FileName: fileName, Rows: i.Rows}
}
