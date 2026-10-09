package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
)

type FormDataRow struct {
	Src         FileRef
	Type        enums.FormDataType
	Key         string
	Value       string
	Description string
	Enabled     bool
}

func (f FormDataRow) IsFile() bool {
	return f.Type == formdatatype.File
}
