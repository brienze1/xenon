package entity

import (
	"strconv"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/scripterrorkind"
)

type ScriptError struct {
	Script  ScriptRef
	Kind    enums.ScriptErrorKind
	Message string
	Line    int
	Column  int
}

func NewScriptError(kind enums.ScriptErrorKind, message string, script ScriptRef) ScriptError {
	return ScriptError{Kind: kind, Message: message, Script: script}
}

func (e ScriptError) At(line, column int) ScriptError {
	e.Line = line
	e.Column = column
	return e
}

func (e ScriptError) HasLocation() bool {
	return e.Line > 0
}

func (e ScriptError) Location() string {
	if !e.HasLocation() {
		return ""
	}
	return "line " + strconv.Itoa(e.Line) + ", column " + strconv.Itoa(e.Column)
}

func (e ScriptError) WithMessage(message string) ScriptError {
	e.Message = message
	return e
}

func NewUnsavedVariablesError(script ScriptRef, reason string) ScriptError {
	return NewScriptError(scripterrorkind.Runtime, "The variable changes of this script could not be saved: "+reason, script)
}
