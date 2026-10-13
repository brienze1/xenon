package dto

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
)

var formDataKeys = []string{"key", "value", "type", "disabled", "description"}

type FormDataRow struct {
	Raw         map[string]any `json:"-"`
	Key         any            `json:"key"`
	Value       any            `json:"value"`
	Src         any            `json:"src"`
	Type        string         `json:"type"`
	Description Description    `json:"description"`
	Disabled    bool           `json:"disabled"`
}

func (f *FormDataRow) UnmarshalJSON(data []byte) error {
	type plain FormDataRow
	if err := json.Unmarshal(data, (*plain)(f)); err != nil {
		return err
	}
	f.Raw = decodeRaw(data)
	return nil
}

func (f *FormDataRow) IsFile() bool {
	return f.Type == formdatatype.File.String()
}

func (f *FormDataRow) ToEntity() entity.FormDataRow {
	row := entity.NewFormDataRow(text(f.Key), text(f.Value)).WithType(f.Type)
	row.Enabled = !f.Disabled
	row.Description = f.Description.Content
	if f.IsFile() {
		row.Value = ""
		row.Src = entity.NewRepickFileRef(f.firstSrc())
	}
	return row
}

func (f *FormDataRow) Passthrough() map[string]any {
	fields := without(f.Raw, formDataKeys...)
	if _, isList := f.Src.([]any); !isList {
		delete(fields, "src")
	}
	return withEntry(fields, "description", f.Description.Passthrough())
}

func (f *FormDataRow) firstSrc() string {
	if sources, isList := f.Src.([]any); isList {
		if len(sources) == 0 {
			return ""
		}
		return text(sources[0])
	}
	return text(f.Src)
}

func NewFormDataRowOf(row entity.FormDataRow, passthrough map[string]any) FormDataRow {
	formRow := FormDataRow{Key: row.Key, Type: row.Type.String(), Disabled: !row.Enabled, Description: newDescription(row.Description, passthrough["description"]), Raw: passthrough}
	if !row.IsFile() {
		formRow.Value = row.Value
		return formRow
	}
	formRow.Src = row.Src.Src
	if sources := asList(passthrough["src"]); len(sources) > 0 && text(sources[0]) == row.Src.Src {
		formRow.Src = sources
	}
	return formRow
}

func (f FormDataRow) MarshalJSON() ([]byte, error) {
	row := newObject().set("key", f.Key).set("type", f.Type)
	if f.Type == formdatatype.File.String() {
		row.set("src", f.Src)
	} else {
		row.set("value", f.Value)
	}
	row.setIf(f.Disabled, "disabled", true).setPresent("description", f.Description.Value())
	return row.merged(f.Raw).MarshalJSON()
}
