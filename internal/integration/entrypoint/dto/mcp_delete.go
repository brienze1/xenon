package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPDelete struct {
	MoveToTrash *bool  `json:"move_to_trash" description:"Move the workspace folder to the trash; default true." used_by:"workspace"`
	Id          string `json:"id" description:"Id of the item, alias, environment or workspace to delete." used_by:"item,alias,environment,workspace" required_for:"item,alias,environment,workspace" validated_as:"local_id@workspace,example_id@item"`
	Domain      string `json:"domain" description:"Cookie domain." used_by:"cookie,cookie_domain" required_for:"cookie,cookie_domain"`
	Name        string `json:"name" description:"Cookie name." used_by:"cookie" required_for:"cookie"`
	Path        string `json:"path" description:"Cookie path, or the absolute folder of the git link." used_by:"cookie,git_link" required_for:"git_link" validated_as:"target_folder@git_link"`
	OwnerId     string `json:"owner_id" description:"Id of the item that owns the OAuth token, or of the request an alias belongs to." used_by:"token,alias" required_for:"token,alias"`
	Source      string `json:"source" description:"History entries to clear; default all." used_by:"history" enum:"all,user,ai"`
}

func (m *MCPDelete) OwnerItemId() string {
	return m.OwnerId
}

func (m *MCPDelete) AliasReference(localId, requestId string) AliasReference {
	return NewAliasReference(localId, requestId, m.Id)
}

func (m *MCPDelete) MovesToTrash() bool {
	return m.MoveToTrash == nil || *m.MoveToTrash
}

func (m *MCPDelete) HasPath() bool {
	return m.Path != ""
}

func (m *MCPDelete) EnvironmentReference(localId string) EnvironmentReference {
	return NewEnvironmentReference(localId, m.Id)
}

func (m *MCPDelete) DeleteCookie(localId string) DeleteCookie {
	return NewDeleteCookie(localId, m.Domain, m.Name, m.Path)
}

func (m *MCPDelete) DeleteCookieDomain(localId string) DeleteCookieDomain {
	return NewDeleteCookieDomain(localId, m.Domain)
}

func (m *MCPDelete) RemoveWorkspace() RemoveWorkspace {
	return NewRemoveWorkspace(m.Id, m.MovesToTrash())
}

func (m *MCPDelete) OAuthTokenReference(localId string) OAuthTokenReference {
	return NewOAuthTokenReference(localId, m.OwnerId)
}

func (m *MCPDelete) UnlinkRepository(localId string) UnlinkRepository {
	return NewUnlinkRepository(localId, m.Path)
}

func (m *MCPDelete) ClearHistory(localId string) ClearHistory {
	return NewClearHistory(localId, m.Source)
}

func (m *MCPDelete) IdTarget() entity.AIActivityTarget {
	return entity.NewAIItemTarget(m.Id)
}

func (m *MCPDelete) DomainTarget() entity.AIActivityTarget {
	return entity.AIActivityTarget{Name: m.Domain}
}

func (m *MCPDelete) OwnerTarget() entity.AIActivityTarget {
	return entity.NewAIItemTarget(m.OwnerId)
}

func (m *MCPDelete) SourceTarget() entity.AIActivityTarget {
	return entity.AIActivityTarget{Name: m.Source}
}

func (m *MCPDelete) ItemId() string {
	return m.Id
}
