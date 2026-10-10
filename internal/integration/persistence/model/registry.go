package model

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type Registry struct {
	Window        json.RawMessage       `json:"window,omitempty"`
	Vault         json.RawMessage       `json:"vault,omitempty"`
	ActiveLocalId string                `json:"active_local_id"`
	Workspaces    []RegisteredWorkspace `json:"workspaces"`
	Version       int                   `json:"version"`
}

func NewRegistry(registry entity.AppRegistry, stored Registry) Registry {
	model := Registry{
		Version:       registry.Version,
		ActiveLocalId: registry.ActiveLocalID,
		Workspaces:    make([]RegisteredWorkspace, 0, len(registry.Workspaces)),
		Window:        stored.Window,
		Vault:         stored.Vault,
	}
	for _, workspace := range registry.Workspaces {
		model.Workspaces = append(model.Workspaces, NewRegisteredWorkspace(workspace))
	}
	return model
}

func (r *Registry) ToEntity() entity.AppRegistry {
	registry := entity.NewAppRegistry()
	registry.ActiveLocalID = r.ActiveLocalId
	for index := range r.Workspaces {
		registry.Workspaces = append(registry.Workspaces, r.Workspaces[index].ToEntity())
	}
	return registry
}
