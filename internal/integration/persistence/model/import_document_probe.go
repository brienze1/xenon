package model

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportDocumentProbe struct {
	Name        string
	IsMalformed bool
}

func (i *ImportDocumentProbe) ToEntity() entity.ImportProbe {
	return entity.ImportProbe{Name: i.Name, IsMalformed: i.IsMalformed}
}

func NewImportDocumentProbe(name string, isMalformed bool) ImportDocumentProbe {
	return ImportDocumentProbe{Name: name, IsMalformed: isMalformed}
}
