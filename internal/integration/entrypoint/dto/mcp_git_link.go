package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/linkmode"
)

type MCPGitLink struct {
	UseRepoWorkspace *bool  `json:"use_repo_workspace"`
	Url              string `json:"url"`
	Branch           string `json:"branch"`
	Subfolder        string `json:"subfolder"`
	AccountId        string `json:"account_id"`
	Path             string `json:"path"`
	Name             string `json:"name"`
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
