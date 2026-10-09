package entity

type FileChange struct {
	Requests []RequestChange
	OwnWrite bool
}
