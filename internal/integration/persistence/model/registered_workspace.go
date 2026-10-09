package model

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const dateTimeLayout = "2006-01-02T15:04:05.000Z"

type RegisteredWorkspace struct {
	LocalId          string `json:"local_id"`
	Path             string `json:"path"`
	DisplayNameCache string `json:"display_name_cache"`
	AddedAt          string `json:"added_at"`
}

func NewRegisteredWorkspace(workspace entity.RegisteredWorkspace) RegisteredWorkspace {
	return RegisteredWorkspace{
		LocalId:          workspace.LocalID,
		Path:             workspace.Path,
		DisplayNameCache: workspace.DisplayNameCache,
		AddedAt:          workspace.AddedAt.UTC().Format(dateTimeLayout),
	}
}

func (r *RegisteredWorkspace) ToEntity() entity.RegisteredWorkspace {
	addedAt, _ := time.Parse(dateTimeLayout, r.AddedAt)
	return entity.RegisteredWorkspace{
		LocalID:          r.LocalId,
		Path:             r.Path,
		DisplayNameCache: r.DisplayNameCache,
		AddedAt:          addedAt,
	}
}
