package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/domain/enums/linkmode"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type gitLinkArguments struct {
	UseRepoWorkspace *bool  `json:"use_repo_workspace"`
	Url              string `json:"url"`
	Branch           string `json:"branch"`
	Subfolder        string `json:"subfolder"`
	AccountId        string `json:"account_id"`
	Path             string `json:"path"`
	Name             string `json:"name"`
}

type gitSettingsArguments struct {
	AutoPull  *bool  `json:"auto_pull"`
	AutoPush  *bool  `json:"auto_push"`
	AccountId string `json:"account_id"`
}

type gitCommitArguments struct {
	Message string   `json:"message"`
	ItemIds []string `json:"item_ids"`
}

type gitPullView struct {
	Session *gitSessionView `json:"session,omitempty"`
	gitStatusView
}

type gitResolveArguments struct {
	Items []gitResolution `json:"items"`
}

type gitResolution struct {
	Id     string `json:"id"`
	Choice string `json:"choice"`
	Field  string `json:"field"`
}

func (t *tools) gitEdit(ctx context.Context, call Call) (Result, error) {
	switch call.Action {
	case "link":
		return t.linkRepository(ctx, call)
	case "clone":
		return t.cloneRepository(ctx, call)
	case "settings":
		return t.saveGitSettings(ctx, call)
	case "commit":
		return t.commitChanges(ctx, call)
	case "pull":
		return t.pullChanges(ctx, call)
	case "resolve":
		return t.resolveConflicts(ctx, call)
	case "finish_pull":
		return t.finishPull(ctx, call)
	default:
		return t.cancelPull(ctx, call)
	}
}

func (t *tools) linkRepository(ctx context.Context, call Call) (Result, error) {
	var arguments gitLinkArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	status, err := t.controllers.SyncLink.Link(ctx, dto.LinkRepository{
		LocalId:    call.LocalId,
		Repository: arguments.Url,
		Branch:     arguments.Branch,
		Subfolder:  arguments.Subfolder,
		AccountID:  arguments.AccountId,
		ClonePath:  arguments.Path,
		Mode:       arguments.linkMode(),
	})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newGitStatusView(*status), Changes: []dto.AIChange{{Kind: aichangekind.Tree.String()}}}, nil
}

func (t *tools) cloneRepository(ctx context.Context, call Call) (Result, error) {
	var arguments gitLinkArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	workspace, err := t.controllers.SyncLink.CloneAsWorkspace(ctx, dto.CloneWorkspace{
		Repository: arguments.Url,
		Branch:     arguments.Branch,
		Subfolder:  arguments.Subfolder,
		AccountID:  arguments.AccountId,
		ClonePath:  arguments.Path,
		Name:       arguments.Name,
	})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: workspace, Target: dto.AIActivityTarget{Id: workspace.LocalId, Name: workspace.Name}, Changes: []dto.AIChange{{Kind: aichangekind.Workspaces.String()}}}, nil
}

func (t *tools) saveGitSettings(ctx context.Context, call Call) (Result, error) {
	var arguments gitSettingsArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	reference := dto.SyncWorkspaceReference{LocalId: call.LocalId}
	current, err := t.controllers.SyncManage.Status(ctx, reference)
	if err != nil {
		return Result{}, err
	}
	if err := t.controllers.SyncManage.SaveSettings(ctx, arguments.merged(call.LocalId, current.Settings)); err != nil {
		return Result{}, err
	}
	return t.gitStatus(ctx, call)
}

func (t *tools) commitChanges(ctx context.Context, call Call) (Result, error) {
	var arguments gitCommitArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	if len(arguments.ItemIds) == 0 {
		changed, err := t.changedItemIds(ctx, call.LocalId)
		if err != nil {
			return Result{}, err
		}
		arguments.ItemIds = changed
	}
	status, err := t.controllers.SyncTransfer.Commit(ctx, dto.CommitSyncChanges{LocalId: call.LocalId, Message: arguments.Message, ItemIDs: arguments.ItemIds})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newGitStatusView(*status)}, nil
}

func (t *tools) pullChanges(ctx context.Context, call Call) (Result, error) {
	pulled, err := t.controllers.SyncTransfer.Pull(ctx, dto.SyncWorkspaceReference{LocalId: call.LocalId})
	if err != nil {
		return Result{}, err
	}
	payload := gitPullView{gitStatusView: newGitStatusView(pulled.Status)}
	if pulled.Session != nil {
		session := newGitSessionView(pulled.Session)
		payload.Session = &session
	}
	return Result{Payload: payload, Changes: []dto.AIChange{{Kind: aichangekind.Tree.String()}}}, nil
}

func (t *tools) resolveConflicts(ctx context.Context, call Call) (Result, error) {
	var arguments gitResolveArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	sessionId, err := t.pullSessionId(ctx, call)
	if err != nil {
		return Result{}, err
	}
	for _, item := range arguments.Items {
		if _, err := t.controllers.SyncConflict.Resolve(ctx, dto.ResolveSyncConflict{SessionId: sessionId, ItemId: item.Id, Resolution: item.Choice, FieldPath: item.Field}); err != nil {
			return Result{}, err
		}
	}
	return t.gitSession(ctx, call)
}

func (t *tools) finishPull(ctx context.Context, call Call) (Result, error) {
	sessionId, err := t.pullSessionId(ctx, call)
	if err != nil {
		return Result{}, err
	}
	status, err := t.controllers.SyncConflict.FinishPull(ctx, dto.SyncSessionReference{SessionId: sessionId})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newGitStatusView(*status), Changes: []dto.AIChange{{Kind: aichangekind.Tree.String()}}}, nil
}

func (t *tools) cancelPull(ctx context.Context, call Call) (Result, error) {
	sessionId, err := t.pullSessionId(ctx, call)
	if err != nil {
		return Result{}, err
	}
	status, err := t.controllers.SyncConflict.CancelPull(ctx, dto.SyncSessionReference{SessionId: sessionId})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newGitStatusView(*status), Changes: []dto.AIChange{{Kind: aichangekind.Tree.String()}}}, nil
}

func (t *tools) changedItemIds(ctx context.Context, localId string) ([]string, error) {
	changes, err := t.controllers.SyncManage.Changes(ctx, dto.SyncWorkspaceReference{LocalId: localId})
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(changes))
	for _, change := range changes {
		ids = append(ids, change.ItemID)
	}
	return ids, nil
}

func (t *tools) pullSessionId(ctx context.Context, call Call) (string, error) {
	session, err := t.controllers.SyncConflict.Session(ctx, dto.SyncWorkspaceReference{LocalId: call.LocalId})
	if err != nil {
		return "", err
	}
	if session == nil {
		return "", invalidArguments(call.Tool, "there is no open pull to "+call.Action)
	}
	return session.ID, nil
}

func (g gitLinkArguments) linkMode() string {
	if g.UseRepoWorkspace == nil {
		return ""
	}
	if *g.UseRepoWorkspace {
		return linkmode.AdoptRepo.String()
	}
	return linkmode.ReplaceWithMine.String()
}

func (g gitSettingsArguments) merged(localId string, current dto.SyncLinkSettings) dto.SaveSyncSettings {
	settings := dto.SaveSyncSettings{LocalId: localId, AccountID: current.AccountID, AutoPull: current.AutoPull, AutoPush: current.AutoPush}
	if g.AccountId != "" {
		settings.AccountID = g.AccountId
	}
	if g.AutoPull != nil {
		settings.AutoPull = *g.AutoPull
	}
	if g.AutoPush != nil {
		settings.AutoPush = *g.AutoPush
	}
	return settings
}
