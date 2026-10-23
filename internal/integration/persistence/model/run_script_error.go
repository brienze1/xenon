package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scripterrorkind"
)

type RunScriptError struct {
	Kind    string       `json:"kind"`
	Message string       `json:"message"`
	Script  RunScriptRef `json:"script"`
	Line    int          `json:"line"`
	Column  int          `json:"column"`
}

func NewRunScriptError(scriptError entity.ScriptError) RunScriptError {
	return RunScriptError{Kind: scriptError.Kind.String(), Message: scriptError.Message, Script: NewRunScriptRef(scriptError.Script), Line: scriptError.Line, Column: scriptError.Column}
}

func (r *RunScriptError) ToEntity() entity.ScriptError {
	return entity.NewScriptError(scripterrorkind.GetScriptErrorKind(r.Kind), r.Message, r.Script.ToEntity()).At(r.Line, r.Column)
}
