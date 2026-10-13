package model

import "github.com/brienze1/xenon/internal/domain/entity"

const localVariablesVersion = 1

type LocalVariables struct {
	Current             map[string]map[string]string `json:"current"`
	ActiveEnvironmentId string                       `json:"activeEnvironmentId"`
	Version             int                          `json:"version"`
}

func NewLocalVariables(local entity.LocalVariables) LocalVariables {
	current := make(map[string]map[string]string, len(local.Current))
	for scope, values := range local.Current {
		if len(values) > 0 {
			current[scope] = values
		}
	}
	return LocalVariables{Version: localVariablesVersion, ActiveEnvironmentId: local.ActiveEnvironmentID, Current: current}
}

func (l *LocalVariables) ToEntity() entity.LocalVariables {
	local := entity.NewLocalVariables().WithActiveEnvironment(l.ActiveEnvironmentId)
	for scope, values := range l.Current {
		local.Current[scope] = values
	}
	return local
}
