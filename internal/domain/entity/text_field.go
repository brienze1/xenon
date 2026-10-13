package entity

import (
	"strconv"
	"strings"
)

type TextField struct {
	Name string
	Text string
}

func NewTextField(name, text string) TextField {
	return TextField{Name: name, Text: text}
}

func (t TextField) References() []Reference {
	return TokenizeReferences(t.Name, t.Text)
}

func fieldName(parts ...string) string {
	name := parts[0]
	for _, part := range parts[1:] {
		name += "." + part
	}
	return name
}

func keyValueFields(prefix string, rows []KeyValue) []TextField {
	fields := make([]TextField, 0, len(rows))
	for index, row := range rows {
		position := strconv.Itoa(index)
		fields = append(fields, NewTextField(fieldName(prefix, position, "key"), row.Key), NewTextField(fieldName(prefix, position, "value"), row.Value))
	}
	return fields
}

func (t TextField) IsLiteral() bool {
	return t.Text != "" && len(t.References()) == 0
}

func (t TextField) HasVaultReference() bool {
	for _, reference := range t.References() {
		if strings.HasPrefix(reference.Name, vaultReferencePrefix) {
			return true
		}
	}
	return false
}
