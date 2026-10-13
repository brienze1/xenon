package entity

import "strings"

type ImportRelocation struct {
	FromID   string
	ToID     string
	FromPath string
	ToPath   string
}

func (i ImportRelocation) Apply(itemId, path string) (string, string) {
	if itemId == i.FromID {
		itemId = i.ToID
	}
	if path == i.FromPath || strings.HasPrefix(path, i.FromPath+importPathSeparator) {
		path = i.ToPath + path[len(i.FromPath):]
	}
	return itemId, path
}

func lastPathSegment(path string) string {
	return path[strings.LastIndex(path, importPathSeparator)+1:]
}
