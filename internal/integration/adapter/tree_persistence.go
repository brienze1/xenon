package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type TreePersistence interface {
	usecase.LoadTree
	usecase.FindTreeItem
	usecase.SaveTreeItem
	usecase.DeleteTreeItem
}
