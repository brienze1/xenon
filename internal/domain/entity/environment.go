package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/envcolor"
)

const productionNameMarker = "prod"

type Environment struct {
	Postman           SourceFields
	Bruno             SourceFields
	ID                string
	Name              string
	Description       string
	Color             enums.EnvColor
	FileName          string
	Error             string
	Variables         []VariableRow
	ConfirmBeforeSend bool
}

// NewEnvironment defaults names containing "prod" to red with confirm-before-send on.
func NewEnvironment(id, name string) Environment {
	environment := Environment{ID: id, Name: strings.TrimSpace(name), Color: envcolor.None, Variables: []VariableRow{}}
	if environment.IsProductionLike() {
		environment.Color = envcolor.Red
		environment.ConfirmBeforeSend = true
	}
	return environment
}

func NewMalformedEnvironment(fileName, reason string) Environment {
	return Environment{
		ID:        errorIdPrefix + fileName,
		Name:      fileName,
		Color:     envcolor.None,
		FileName:  fileName,
		Error:     reason,
		Variables: []VariableRow{},
	}
}

func (e Environment) IsProductionLike() bool {
	return strings.Contains(strings.ToLower(e.Name), productionNameMarker)
}

func (e Environment) IsMalformed() bool {
	return e.Error != ""
}

func (e Environment) HasID(id string) bool {
	return !e.IsMalformed() && e.ID == id
}

func (e Environment) HasName(name string) bool {
	return strings.EqualFold(strings.TrimSpace(e.Name), strings.TrimSpace(name))
}

func (e Environment) HasValidName() bool {
	return isValidItemName(e.Name)
}

func (e Environment) HasValidColor() bool {
	return e.Color != envcolor.Unknown
}

func (e Environment) WithSettings(settings Environment) Environment {
	e.Name = strings.TrimSpace(settings.Name)
	e.Description = settings.Description
	e.Color = settings.Color
	e.ConfirmBeforeSend = settings.ConfirmBeforeSend
	return e
}

func (e Environment) WithVariables(rows []VariableRow) Environment {
	e.Variables = append([]VariableRow{}, rows...)
	return e
}

func (e Environment) WithFileName(fileName string) Environment {
	e.FileName = fileName
	return e
}

func (e Environment) Duplicated(id, name string) Environment {
	e.ID = id
	e.Name = name
	e.FileName = ""
	return e.WithVariables(e.Variables)
}

func (e Environment) ScopeRef() ScopeRef {
	return NewEnvironmentScopeRef(e.ID)
}

func (e Environment) sortKey() string {
	return strings.ToLower(e.Name) + "\x00" + e.FileName
}

func (e Environment) LogValue() any {
	return map[string]any{"id": e.ID, "name": e.Name, "color": e.Color, "variables": len(e.Variables), "error": e.Error}
}
