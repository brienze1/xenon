package usecase

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/conflictresolution"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

const orderField = "order"

var knownStructuralFields = map[string]bool{
	"xenon": true, "kind": true, "id": true, "name": true, "description": true, "order": true, "auth": true,
	"headers": true, "scripts": true, "variables": true, "settings": true, "sync": true, "docs": true,
	"color": true, "confirmBeforeSend": true, "": true,
}

type structuralMerger struct {
	merge         entity.StructuralMerge
	isEnvironment bool
}

func MergeStructuralFile(kind enums.ItemKind, base, mine, theirs *entity.DocumentNode) entity.StructuralMerge {
	merger := &structuralMerger{isEnvironment: kind == itemkind.Environment}
	merger.merge.Result = merger.value("", base, mine, theirs)
	return merger.merge
}

func (s *structuralMerger) value(path string, base, mine, theirs *entity.DocumentNode) *entity.DocumentNode {
	switch {
	case mine.Equals(theirs):
		return mine
	case base.Equals(mine):
		return theirs
	case base.Equals(theirs):
		return mine
	case mine.IsMapping() && theirs.IsMapping() && (base == nil || base.IsMapping()):
		return s.mapping(path, base, mine, theirs)
	}
	if path == orderField {
		if merged, ok := mergeOrder(base, mine, theirs); ok {
			return merged
		}
	}
	if merged, ok := s.rows(path, base, mine, theirs); ok {
		return merged
	}
	return s.conflict(path, mine, theirs)
}

func (s *structuralMerger) mapping(path string, base, mine, theirs *entity.DocumentNode) *entity.DocumentNode {
	result := entity.NewMappingNode()
	for _, key := range unionInOrder(theirs.Keys, mine.Keys) {
		if merged := s.value(childPath(path, key), base.Field(key), mine.Field(key), theirs.Field(key)); merged != nil {
			result.Set(key, merged)
		}
	}
	return result
}

func (s *structuralMerger) rows(path string, base, mine, theirs *entity.DocumentNode) (*entity.DocumentNode, bool) {
	mineKeys, mineRows, mineKeyed := mine.KeyedRows()
	theirKeys, theirRows, theirsKeyed := theirs.KeyedRows()
	_, baseRows, baseKeyed := base.KeyedRows()
	if base == nil {
		baseRows, baseKeyed = map[string]*entity.DocumentNode{}, true
	}
	if !mineKeyed || !theirsKeyed || !baseKeyed {
		return nil, false
	}
	items := make([]*entity.DocumentNode, 0, len(theirKeys)+len(mineKeys))
	for _, key := range unionInOrder(theirKeys, mineKeys) {
		if merged := s.value(path+"["+key+"]", baseRows[key], mineRows[key], theirRows[key]); merged != nil {
			items = append(items, merged)
		}
	}
	return entity.NewSequenceNode(items), true
}

func (s *structuralMerger) conflict(path string, mine, theirs *entity.DocumentNode) *entity.DocumentNode {
	if s.isEnvironment {
		s.merge.ItemConflict = true
		return mine
	}
	if !knownStructuralFields[topField(path)] {
		s.merge.Notes = append(s.merge.Notes, "kept their "+path)
		return theirs
	}
	s.merge.FieldConflicts = append(s.merge.FieldConflicts, entity.FieldConflict{Path: path, Mine: mine, Theirs: theirs, Resolution: conflictresolution.Unresolved})
	return mine
}

func mergeOrder(base, mine, theirs *entity.DocumentNode) (*entity.DocumentNode, bool) {
	mineItems, mineScalars := mine.ScalarItems()
	theirItems, theirsScalars := theirs.ScalarItems()
	baseItems, baseScalars := base.ScalarItems()
	if !mineScalars || !theirsScalars || (base != nil && !baseScalars) {
		return nil, false
	}
	inBase, inMine, inTheirs := toSet(baseItems), toSet(mineItems), toSet(theirItems)
	seen := map[string]bool{}
	items := make([]*entity.DocumentNode, 0, len(theirItems)+len(mineItems))
	for _, item := range theirs.Items {
		if !seen[item.Value] && (inMine[item.Value] || !inBase[item.Value]) {
			seen[item.Value] = true
			items = append(items, item)
		}
	}
	for _, item := range mine.Items {
		if !seen[item.Value] && !inTheirs[item.Value] && !inBase[item.Value] {
			seen[item.Value] = true
			items = append(items, item)
		}
	}
	return entity.NewSequenceNode(items), true
}

func unionInOrder(first, second []string) []string {
	seen := map[string]bool{}
	union := make([]string, 0, len(first)+len(second))
	for _, value := range append(append([]string{}, first...), second...) {
		if !seen[value] {
			seen[value] = true
			union = append(union, value)
		}
	}
	return union
}

func toSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func childPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

func topField(path string) string {
	field, _, _ := strings.Cut(strings.SplitN(path, "[", 2)[0], ".")
	return field
}
