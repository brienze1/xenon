package usecase

import (
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
)

const importedSuffix = " (imported)"

func ResolveClashName(name string, taken []string) string {
	candidate := name + importedSuffix
	for suffix := 2; isNameTaken(candidate, taken); suffix++ {
		candidate = name + " (imported " + strconv.Itoa(suffix) + ")"
	}
	return candidate
}

func ResolveImportClashes(nodes []entity.ImportNode, existing []entity.TreeNode, resolution enums.ClashResolution) ([]entity.ImportNode, []entity.TreeNode) {
	kept := make([]entity.ImportNode, 0, len(nodes))
	replaced := make([]entity.TreeNode, 0)
	taken := existingNames(existing)
	for _, node := range nodes {
		match, clashes := sameNamed(node.Name(), existing)
		switch {
		case !clashes:
			kept = append(kept, node)
		case resolution == clashresolution.Replace:
			kept = append(kept, node)
			replaced = append(replaced, match)
		case resolution == clashresolution.KeepBoth:
			renamed := node.Renamed(ResolveClashName(node.Name(), taken))
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

func isNameTaken(candidate string, taken []string) bool {
	for _, name := range taken {
		if strings.EqualFold(strings.TrimSpace(name), candidate) {
			return true
		}
	}
	return false
}
