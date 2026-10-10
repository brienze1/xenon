package usecase

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

func PlanImportWrite(placement entity.ImportPlacement, siblings entity.TreeNode, replaced []entity.TreeNode, collection *entity.Collection, variables []entity.VariableRow, resolution enums.ClashResolution) entity.ImportWrite {
	replacedFiles := make([]string, 0, len(replaced))
	for _, node := range replaced {
		replacedFiles = append(replacedFiles, node.FileName)
	}
	named := NameImportFiles(placement.Nodes, siblings.EntriesWithout(replacedFiles))
	replacements, appended := importedFileNames(named, replaced)
	write := entity.ImportWrite{Nodes: named, Replaced: replaced, Order: siblings.OrderImporting(replacements, appended), Merge: placement.Merge}
	if placement.Parent != nil {
		parent := placement.Parent.WithOrder(write.Order)
		write.Parent = &parent
	}
	if collection == nil {
		return write
	}
	rows, kept := MergeImportVariables(collection.Variables, variables, resolution)
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
		if existing, found := replacedBy(node, replaced); found {
			replacements[existing.FileName] = node.FileName
			continue
		}
		appended = append(appended, node.FileName)
	}
	return replacements, appended
}

func replacedBy(node entity.ImportNode, replaced []entity.TreeNode) (entity.TreeNode, bool) {
	for _, existing := range replaced {
		if strings.EqualFold(strings.TrimSpace(existing.Name), strings.TrimSpace(node.Name())) {
			return existing, true
		}
	}
	return entity.TreeNode{}, false
}
