package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportItemResult struct {
	Key     string `json:"key"`
	Status  string `json:"status"`
	LocalId string `json:"local_id"`
	Id      string `json:"id"`
	Name    string `json:"name"`
}

func NewImportItemResult(item entity.ImportItemResult) ImportItemResult {
	return ImportItemResult{Key: item.Key, Status: item.Status.String(), LocalId: item.LocalID, Id: item.ID, Name: item.Name}
}
