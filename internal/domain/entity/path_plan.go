package entity

import (
	"path"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/domain/enums/pathaction"
)

const (
	collectionFileName = "collection.yaml"
	folderFileName     = "folder.yaml"
)

type PathPlan struct {
	Path   string
	Base   string
	Mine   string
	Theirs string
	Action enums.PathAction
}

func NewPathPlan(filePath, base, mine, theirs string) PathPlan {
	plan := PathPlan{Path: filePath, Base: base, Mine: mine, Theirs: theirs}
	plan.Action = plan.classify()
	return plan
}

func (p PathPlan) StructuralKind() enums.ItemKind {
	switch {
	case p.Path == workspaceFileName:
		return itemkind.Workspace
	case path.Base(p.Path) == collectionFileName:
		return itemkind.Collection
	case path.Base(p.Path) == folderFileName:
		return itemkind.Folder
	case path.Dir(p.Path) == environmentsFolder && strings.HasSuffix(p.Path, ".yaml"):
		return itemkind.Environment
	default:
		return itemkind.Unknown
	}
}

func (p PathPlan) IsConflict() bool {
	return p.Action == pathaction.BothModified || p.Action == pathaction.DeleteModify
}

func (p PathPlan) Result() string {
	if p.Action == pathaction.TakeTheirs {
		return p.Theirs
	}
	return p.Mine
}

func (p PathPlan) classify() enums.PathAction {
	switch {
	case p.Mine == p.Theirs:
		return pathaction.Keep
	case p.Mine == p.Base:
		return pathaction.TakeTheirs
	case p.Theirs == p.Base:
		return pathaction.TakeMine
	case p.Mine == "" || p.Theirs == "":
		return pathaction.DeleteModify
	case p.StructuralKind() != itemkind.Unknown:
		return pathaction.MergeStructural
	default:
		return pathaction.BothModified
	}
}

func (p PathPlan) AnyVersion() string {
	for _, version := range []string{p.Theirs, p.Base, p.Mine} {
		if version != "" {
			return version
		}
	}
	return ""
}
