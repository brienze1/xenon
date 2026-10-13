package entity

import "github.com/brienze1/xenon/internal/domain/enums/formdatatype"

type ApiField struct {
	Name        string
	Description string
	Example     string
	Binary      bool
}

func NewApiField(row KeyValue) ApiField {
	return ApiField{Name: row.Key, Description: row.Description, Example: row.Value}
}

func NewApiFormField(row FormDataRow) ApiField {
	if row.IsFile() {
		return ApiField{Name: row.Key, Description: row.Description, Binary: true}
	}
	return ApiField{Name: row.Key, Description: row.Description, Example: row.Value}
}

func (a ApiField) KeyValue() KeyValue {
	row := NewKeyValue(a.Name, a.Example)
	row.Description = a.Description
	return row
}

func (a ApiField) FormDataRow(current []FormDataRow) FormDataRow {
	row := NewFormDataRow(a.Name, a.Example)
	row.Description = a.Description
	if !a.Binary {
		return row
	}
	row.Type = formdatatype.File
	row.Value = ""
	for _, existing := range current {
		if existing.IsFile() && existing.Key == a.Name {
			row.Src = existing.Src
		}
	}
	return row
}
