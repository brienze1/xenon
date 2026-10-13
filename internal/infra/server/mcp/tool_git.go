package mcp

import (
	"context"
	"net/url"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums/conflictkind"
	"github.com/brienze1/xenon/internal/domain/enums/conflictresolution"
	"github.com/brienze1/xenon/internal/domain/enums/syncstate"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

const gitPathSeparator = ", "

type gitStatusView struct {
	Error      *dto.SyncError `json:"error,omitempty"`
	LastSyncAt *string        `json:"last_sync_at,omitempty"`
	State      string         `json:"state"`
	Branch     string         `json:"branch,omitempty"`
	Remote     string         `json:"remote,omitempty"`
	Subfolder  string         `json:"subfolder,omitempty"`
	AccountId  string         `json:"account_id,omitempty"`
	Notice     string         `json:"notice,omitempty"`
	Ahead      int            `json:"ahead,omitempty"`
	Behind     int            `json:"behind,omitempty"`
	Changes    int            `json:"changes,omitempty"`
	Linked     bool           `json:"linked"`
	AutoPull   bool           `json:"auto_pull,omitempty"`
	AutoPush   bool           `json:"auto_push,omitempty"`
}

type gitChangeLine struct {
	Id     string `json:"id"`
	Path   string `json:"path,omitempty"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Change string `json:"change"`
}

type gitSessionView struct {
	Id    string            `json:"id,omitempty"`
	Items []gitConflictLine `json:"items"`
}

type gitConflictLine struct {
	Id         string             `json:"id"`
	Name       string             `json:"name"`
	Kind       string             `json:"kind"`
	Conflict   string             `json:"conflict"`
	Path       string             `json:"path,omitempty"`
	Resolution string             `json:"resolution,omitempty"`
	Choices    []string           `json:"choices"`
	Fields     []gitConflictField `json:"fields,omitempty"`
}

type gitConflictField struct {
	Path       string `json:"path"`
	Resolution string `json:"resolution,omitempty"`
}

func newGitStatusView(status dto.SyncStatus) gitStatusView {
	return gitStatusView{
		Error:      status.LastError,
		LastSyncAt: status.LastSyncAt,
		State:      status.State,
		Branch:     status.Settings.Branch,
		Remote:     withoutCredentials(status.Settings.Repository),
		Subfolder:  status.Settings.Subfolder,
		AccountId:  status.Settings.AccountID,
		Notice:     status.Notice,
		Ahead:      status.Ahead,
		Behind:     status.Behind,
		Changes:    status.LocalChanges,
		Linked:     status.State != syncstate.NotLinked.String(),
		AutoPull:   status.Settings.AutoPull,
		AutoPush:   status.Settings.AutoPush,
	}
}

func newGitSessionView(session *dto.SyncSession) gitSessionView {
	if session == nil {
		return gitSessionView{Items: []gitConflictLine{}}
	}
	items := make([]gitConflictLine, 0, len(session.Conflicts))
	for _, conflict := range session.Conflicts {
		items = append(items, newGitConflictLine(conflict))
	}
	return gitSessionView{Id: session.ID, Items: items}
}

func newGitConflictLine(conflict dto.SyncConflictItem) gitConflictLine {
	fields := make([]gitConflictField, 0, len(conflict.Fields))
	for _, field := range conflict.Fields {
		fields = append(fields, gitConflictField{Path: field.Path, Resolution: unresolvedAsEmpty(field.Resolution)})
	}
	return gitConflictLine{
		Id:         conflict.ItemID,
		Name:       conflict.ItemName,
		Kind:       conflict.ItemKind,
		Conflict:   conflict.Kind,
		Path:       conflict.Path,
		Resolution: unresolvedAsEmpty(conflict.Resolution),
		Choices:    conflictChoices(conflict.Kind),
		Fields:     fields,
	}
}

func conflictChoices(kind string) []string {
	switch kind {
	case conflictkind.Field.String():
		return []string{conflictresolution.Mine.String(), conflictresolution.Theirs.String()}
	case conflictkind.DeleteModify.String():
		return []string{conflictresolution.Keep.String(), conflictresolution.Delete.String()}
	default:
		return []string{conflictresolution.Mine.String(), conflictresolution.Theirs.String(), conflictresolution.Both.String()}
	}
}

func unresolvedAsEmpty(resolution string) string {
	if resolution == conflictresolution.Unresolved.String() {
		return ""
	}
	return resolution
}

func withoutCredentials(repository string) string {
	parsed, err := url.Parse(repository)
	if err != nil || parsed.User == nil {
		return repository
	}
	parsed.User = nil
	return parsed.String()
}

func (t *tools) git(ctx context.Context, call Call) (Result, error) {
	switch call.Action {
	case "changes":
		return t.gitChanges(ctx, call)
	case "session":
		return t.gitSession(ctx, call)
	default:
		return t.gitStatus(ctx, call)
	}
}

func (t *tools) gitStatus(ctx context.Context, call Call) (Result, error) {
	status, err := t.controllers.SyncManage.Status(ctx, dto.SyncWorkspaceReference{LocalId: call.LocalId})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newGitStatusView(*status)}, nil
}

func (t *tools) gitChanges(ctx context.Context, call Call) (Result, error) {
	changes, err := t.controllers.SyncManage.Changes(ctx, dto.SyncWorkspaceReference{LocalId: call.LocalId})
	if err != nil {
		return Result{}, err
	}
	lines := make([]gitChangeLine, 0, len(changes))
	for _, change := range changes {
		lines = append(lines, gitChangeLine{Id: change.ItemID, Path: strings.Join(change.Paths, gitPathSeparator), Name: change.ItemName, Kind: change.ItemKind, Change: change.Change})
	}
	return Result{Payload: List[gitChangeLine]{Items: lines}}, nil
}

func (t *tools) gitSession(ctx context.Context, call Call) (Result, error) {
	session, err := t.controllers.SyncConflict.Session(ctx, dto.SyncWorkspaceReference{LocalId: call.LocalId})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newGitSessionView(session)}, nil
}
