package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type ImportItemResult struct {
	Status  enums.ImportStatus
	Key     string
	LocalID string
	ID      string
	Name    string
}
