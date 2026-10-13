package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
)

func ResolveImportClashes(nodes []entity.ImportNode, existing []entity.TreeNode, resolution enums.ClashResolution) ([]entity.ImportNode, []entity.TreeNode) {
	kept := make([]entity.ImportNode, 0, len(nodes))
	replaced := make([]entity.TreeNode, 0)
	taken := existingNames(existing)
	for _, node := range nodes {
		match, clashes := entity.NamedTreeNode(existing, node.Name())
		switch {
		case !clashes:
			kept = append(kept, node)
		case resolution == clashresolution.Replace:
			kept = append(kept, node)
			replaced = append(replaced, match)
		case resolution == clashresolution.KeepBoth:
			renamed := node.Renamed(entity.ImportedName(node.Name(), taken))
			taken = append(taken, renamed.Name())
			kept = append(kept, renamed)
		}
	}
	return kept, replaced
}

func existingNames(existing []entity.TreeNode) []string {
	names := make([]string, 0, len(existing))
	for _, node := range existing {
		names = append(names, node.Name)
	}
	return names
}
