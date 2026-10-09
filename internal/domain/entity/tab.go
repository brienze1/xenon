package entity

type Tab struct {
	Draft        *Request
	TabID        string
	RequestID    string
	BaseRevision string
	Pinned       bool
}
