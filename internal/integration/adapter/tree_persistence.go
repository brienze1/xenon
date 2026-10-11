package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type TreePersistence interface {
	usecase.LoadTree
	usecase.FindTreeItem
	usecase.CreateTreeItem
	usecase.RenameTreeItem
	usecase.DuplicateTreeItem
	usecase.MoveTreeItem
	usecase.SaveTreeItem
	usecase.DeleteTreeItem
	usecase.InspectFileChange
}
