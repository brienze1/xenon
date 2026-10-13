package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type deleteArguments struct {
	MoveToTrash *bool  `json:"move_to_trash"`
	Id          string `json:"id"`
	Domain      string `json:"domain"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	OwnerId     string `json:"owner_id"`
	Source      string `json:"source"`
}

type deletedView struct {
	Deleted string `json:"deleted"`
	Id      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Path    string `json:"path,omitempty"`
}

func (a deleteArguments) movesToTrash() bool {
	return a.MoveToTrash == nil || *a.MoveToTrash
}

func (t *tools) delete(ctx context.Context, call Call) (Result, error) {
	var arguments deleteArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	target, kind, err := t.deleteTarget(ctx, call, arguments)
	if err != nil {
		return Result{}, err
	}
	result := Result{Payload: deletedView{Deleted: call.Action, Id: target.Id, Name: target.Name, Path: target.Path}, Target: target}
	if kind != "" {
		result.Changes = []dto.AIChange{{Kind: kind}}
	}
	return result, nil
}

func (t *tools) deleteTarget(ctx context.Context, call Call, arguments deleteArguments) (dto.AIActivityTarget, string, error) {
	switch call.Action {
	case "item":
		return t.deleteItem(ctx, call, arguments)
	case "environment":
		_, err := t.controllers.EnvironmentManage.Delete(ctx, dto.EnvironmentReference{LocalId: call.LocalId, Id: arguments.Id})
		return dto.AIActivityTarget{Id: arguments.Id}, aichangekind.Variables.String(), err
	case "cookie":
		err := t.controllers.Cookie.Delete(ctx, dto.DeleteCookie{LocalId: call.LocalId, Domain: arguments.Domain, Name: arguments.Name, Path: arguments.Path})
		return dto.AIActivityTarget{Name: arguments.Name, Path: arguments.Domain + arguments.Path}, aichangekind.Cookies.String(), err
	case "cookie_domain":
		err := t.controllers.Cookie.DeleteDomain(ctx, dto.DeleteCookieDomain{LocalId: call.LocalId, Domain: arguments.Domain})
		return dto.AIActivityTarget{Name: arguments.Domain}, aichangekind.Cookies.String(), err
	case "cookies":
		return dto.AIActivityTarget{}, aichangekind.Cookies.String(), t.controllers.Cookie.Clear(ctx, dto.CookieJarReference{LocalId: call.LocalId})
	case "workspace":
		_, err := t.controllers.WorkspaceRemove.Handle(ctx, dto.RemoveWorkspace{LocalId: arguments.Id, MoveToTrash: arguments.movesToTrash()})
		return dto.AIActivityTarget{Id: arguments.Id}, aichangekind.Workspaces.String(), err
	case "token":
		return dto.AIActivityTarget{Id: arguments.OwnerId}, "", t.controllers.OAuth.Clear(ctx, dto.OAuthTokenReference{LocalId: call.LocalId, OwnerId: arguments.OwnerId})
	case "git_link":
		return t.unlinkRepository(ctx, call, arguments)
	case "history":
		return dto.AIActivityTarget{Name: arguments.Source}, aichangekind.History.String(), t.controllers.History.Clear(ctx, dto.ClearHistory{LocalId: call.LocalId, Source: arguments.Source})
	default:
		return dto.AIActivityTarget{}, "", t.controllers.Console.Clear(ctx)
	}
}

func (t *tools) deleteItem(ctx context.Context, call Call, arguments deleteArguments) (dto.AIActivityTarget, string, error) {
	item, err := t.findWorkspaceItem(ctx, call.LocalId, arguments.Id)
	if err != nil {
		return dto.AIActivityTarget{}, "", err
	}
	if item.isExample() {
		err = t.controllers.ExampleManage.Delete(ctx, dto.ExampleReference{LocalId: call.LocalId, RequestId: item.RequestId, ExampleId: item.Id})
	} else {
		_, err = t.controllers.TreeManage.Delete(ctx, dto.TreeItemReference{LocalId: call.LocalId, Id: item.Id})
	}
	return item.target(), aichangekind.Tree.String(), err
}

func (t *tools) unlinkRepository(ctx context.Context, call Call, arguments deleteArguments) (dto.AIActivityTarget, string, error) {
	if arguments.Path == "" {
		return dto.AIActivityTarget{}, "", invalidArguments(call.Tool, "path (an empty folder that keeps the workspace files) is required to unlink")
	}
	summary, err := t.controllers.SyncLink.Unlink(ctx, dto.UnlinkRepository{LocalId: call.LocalId, TargetFolder: arguments.Path})
	if err != nil {
		return dto.AIActivityTarget{}, "", err
	}
	return dto.AIActivityTarget{Id: summary.LocalId, Name: summary.Name, Path: summary.Path}, aichangekind.Workspaces.String(), nil
}
