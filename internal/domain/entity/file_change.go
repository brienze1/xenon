package entity

type FileChange struct {
	Requests []RequestChange
	OwnWrite bool
}

func (f FileChange) IsOwnWrite() bool {
	return f.OwnWrite
}
