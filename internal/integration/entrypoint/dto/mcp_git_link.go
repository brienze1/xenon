package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/linkmode"
)

type MCPGitLink struct {
	UseRepoWorkspace *bool  `json:"use_repo_workspace" description:"Use the workspace already in the repository instead of replacing it with this one." used_by:"link" validated_as:"mode"`
	Url              string `json:"url" description:"Repository URL." required_for:"link,clone" validated_as:"repository"`
	Branch           string `json:"branch" description:"Branch to sync." required_for:"link"`
	Subfolder        string `json:"subfolder" description:"Folder inside the repository that holds the workspace."`
	AccountId        string `json:"account_id" description:"Git account id, from the Sync settings." required_for:"link,clone"`
	Path             string `json:"path" description:"Absolute folder to clone into." validated_as:"clone_path"`
	Name             string `json:"name" description:"Name of the cloned workspace, at most 200 characters." used_by:"clone"`
}

func (m *MCPGitLink) LinkRepository(localId string) LinkRepository {
	return NewLinkRepository(localId, m.Url, m.Branch, m.Subfolder, m.AccountId, m.Path, m.linkMode())
}

func (m *MCPGitLink) CloneWorkspace() CloneWorkspace {
	return NewCloneWorkspace(m.Url, m.Branch, m.Subfolder, m.AccountId, m.Path, m.Name)
}

func (m *MCPGitLink) linkMode() string {
	if m.UseRepoWorkspace == nil {
		return ""
	}
	if *m.UseRepoWorkspace {
		return linkmode.AdoptRepo.String()
	}
	return linkmode.ReplaceWithMine.String()
}

func (m MCPGitLink) LogValue() any {
	return map[string]any{
		"use_repo_workspace": m.UseRepoWorkspace,
		"url":                entity.RemoteWithoutCredentials(m.Url),
		"branch":             m.Branch,
		"subfolder":          m.Subfolder,
		"account_id":         m.AccountId,
		"path":               m.Path,
		"name":               m.Name,
	}
}
