package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
)

type AIChange struct {
	Kind      enums.AIChangeKind
	RequestID string
	Revision  string
}

func (a AIChange) IsTree() bool {
	return a.Kind == aichangekind.Tree
}

func (a AIChange) IsRequest() bool {
	return a.Kind == aichangekind.Request
}

func (a AIChange) IsVariables() bool {
	return a.Kind == aichangekind.Variables
}

func (a AIChange) IsHistory() bool {
	return a.Kind == aichangekind.History
}

func (a AIChange) IsCookies() bool {
	return a.Kind == aichangekind.Cookies
}

func (a AIChange) IsWorkspaces() bool {
	return a.Kind == aichangekind.Workspaces
}

func (a AIChange) RequestChange() RequestChange {
	return RequestChange{RequestID: a.RequestID, Revision: a.Revision}
}

func (a AIChange) LogValue() any {
	return map[string]any{"kind": a.Kind, "requestId": a.RequestID}
}
