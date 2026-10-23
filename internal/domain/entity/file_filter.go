package entity

type FileFilter struct {
	DisplayName string
	Pattern     string
}

func NewFileFilter(displayName, pattern string) FileFilter {
	return FileFilter{DisplayName: displayName, Pattern: pattern}
}
