package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type TreeDocument struct {
	Relative string
	Document entity.ImportDocument
}

func NewTreeDocument(document entity.ImportDocument, relative string) TreeDocument {
	return TreeDocument{Document: document, Relative: relative}
}
