package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

func PlanImportWrite(placement entity.ImportPlacement, siblings entity.TreeNode, replaced []entity.TreeNode, collection *entity.Collection, variables []entity.VariableRow, resolution enums.ClashResolution) entity.ImportWrite {
	replacements, appended := importedFileNames(placement.Nodes, replaced)
	write := entity.ImportWrite{Nodes: placement.Nodes, Replaced: replaced, Order: siblings.OrderImporting(replacements, appended), Merge: placement.Merge, Wrapped: placement.Wrapped}
	if placement.Parent != nil {
		parent := placement.Parent.WithOrder(write.Order)
		write.Parent = &parent
	}
	if collection == nil {
		return write
	}
	rows, kept := entity.MergeImportedVariables(collection.Variables, variables, resolution)
	write.KeptVariables = kept
	write.AddedVariables = len(rows) - len(collection.Variables)
	if write.Parent != nil && write.Parent.ID == collection.ID {
		parent := write.Parent.WithVariables(rows)
		write.Parent = &parent
		return write
	}
	updated := collection.WithVariables(rows)
	write.Collection = &updated
	return write
}

func importedFileNames(named []entity.ImportNode, replaced []entity.TreeNode) (map[string]string, []string) {
	replacements := map[string]string{}
	appended := make([]string, 0, len(named))
	for _, node := range named {
		if existing, found := entity.NamedTreeNode(replaced, node.Name()); found {
			replacements[existing.FileName] = node.FileName
			continue
		}
		appended = append(appended, node.FileName)
	}
	return replacements, appended
}
