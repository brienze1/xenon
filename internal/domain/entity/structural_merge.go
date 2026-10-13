package entity

type StructuralMerge struct {
	Result         *DocumentNode
	FieldConflicts []FieldConflict
	Notes          []string
	ItemConflict   bool
}

func (s StructuralMerge) HasConflicts() bool {
	return s.ItemConflict || len(s.FieldConflicts) > 0
}
