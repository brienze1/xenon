package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptCheck struct {
	Errors []ScriptSyntaxError `json:"errors"`
}

func NewScriptCheck(errors []entity.ScriptError) ScriptCheck {
	result := ScriptCheck{Errors: make([]ScriptSyntaxError, 0, len(errors))}
	for _, scriptError := range errors {
		result.Errors = append(result.Errors, NewScriptSyntaxError(scriptError))
	}
	return result
}
