package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type TreeDocument struct {
	Document entity.ImportDocument
	Relative string
}
