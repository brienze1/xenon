package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptCheck struct {
	Errors []ScriptSyntaxError `json:"errors"`
}

type ScriptSyntaxError struct {
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
}

func NewScriptCheck(errors []entity.ScriptError) ScriptCheck {
	result := ScriptCheck{Errors: make([]ScriptSyntaxError, 0, len(errors))}
	for _, scriptError := range errors {
		result.Errors = append(result.Errors, ScriptSyntaxError{Message: scriptError.Message, Line: scriptError.Line, Column: scriptError.Column})
	}
	return result
}
