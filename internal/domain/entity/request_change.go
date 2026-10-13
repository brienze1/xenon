package entity

type RequestChange struct {
	RequestID string
	Revision  string
	Deleted   bool
}

func (r RequestChange) BlocksSave(baseRevision string, force bool) bool {
	return !force && r.Revision != baseRevision
}
