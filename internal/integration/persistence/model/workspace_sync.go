package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type WorkspaceSync struct {
	Repository string `yaml:"repository"`
	Branch     string `yaml:"branch"`
	Subfolder  string `yaml:"subfolder"`
}

func NewWorkspaceSync(settings entity.SyncSettings) WorkspaceSync {
	return WorkspaceSync{Repository: settings.Repository, Branch: settings.Branch, Subfolder: settings.Subfolder}
}

func (w *WorkspaceSync) ToEntity() entity.SyncSettings {
	return entity.NewSyncSettings(w.Repository, w.Branch, w.Subfolder)
}

func (w *WorkspaceSync) Fields() []yamlnode.Field {
	return []yamlnode.Field{yamlnode.Object("sync",
		yamlnode.String("repository", w.Repository),
		yamlnode.String("branch", w.Branch),
		yamlnode.String("subfolder", w.Subfolder),
	)}
}
