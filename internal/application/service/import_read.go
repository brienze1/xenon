package service

import (
	"strconv"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

type importRead struct {
	secrets  string
	items    []entity.ImportItem
	warnings []entity.ImportWarning
}

func (i *importRead) addTree(kind enums.ImportItemKind, root entity.ImportNode, warnings []entity.ImportWarning) {
	item := entity.NewImportItem(i.nextKey(), kind, root, warnings)
	usages := usecase.ScanPmUsage(item.Scripts())
	i.items = append(i.items, item.WithPmUsages(usages).WithWarnings(usecase.ListPmUsageWarnings(usages)))
}

func (i *importRead) addCollection(kind enums.ImportItemKind, collection entity.ImportCollection) {
	i.addTree(kind, collection.Root, collection.Warnings)
	i.items[len(i.items)-1] = i.items[len(i.items)-1].WithCollectionSecrets(collection.Secrets)
}

func (i *importRead) addVariables(kind enums.ImportItemKind, variables entity.ImportVariables) {
	item := entity.NewVariablesImportItem(i.nextKey(), kind, variables)
	i.items = append(i.items, item.WithWarnings(variables.Warnings))
}

func (i *importRead) addSourcedTree(source, root entity.ImportNode) {
	i.addTree(entity.ImportItemKindOf(root), root, []entity.ImportWarning{})
	i.items[len(i.items)-1] = i.items[len(i.items)-1].FromSource(source.ID())
}

func (i *importRead) nextKey() string {
	return strconv.Itoa(len(i.items))
}
