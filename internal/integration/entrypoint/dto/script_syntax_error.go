package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptSyntaxError struct {
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
}

func NewScriptSyntaxError(scriptError entity.ScriptError) ScriptSyntaxError {
	return ScriptSyntaxError{Message: scriptError.Message, Line: scriptError.Line, Column: scriptError.Column}
}
