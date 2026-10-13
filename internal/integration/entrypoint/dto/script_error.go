package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptError struct {
	Line    *int      `json:"line,omitempty"`
	Column  *int      `json:"column,omitempty"`
	Kind    string    `json:"kind"`
	Message string    `json:"message"`
	Script  ScriptRef `json:"script"`
}

func NewScriptError(scriptError entity.ScriptError) ScriptError {
	result := ScriptError{Kind: scriptError.Kind.String(), Message: scriptError.Message, Script: NewScriptRef(scriptError.Script)}
	if scriptError.HasLocation() {
		line, column := scriptError.Line, scriptError.Column
		result.Line, result.Column = &line, &column
	}
	return result
}
