package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPDelete struct {
	MoveToTrash *bool  `json:"move_to_trash"`
	Id          string `json:"id"`
	Domain      string `json:"domain"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	OwnerId     string `json:"owner_id"`
	Source      string `json:"source"`
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
