package entity

type ImportProgress struct {
	SourceID    string
	CurrentItem string
	Done        int
	Total       int
}
