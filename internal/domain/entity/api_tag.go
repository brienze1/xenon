package entity

import "strings"

const TagSeparator = " / "

type ApiTag struct {
	Name        string
	Description string
}

func NewApiTag(folders []Collection) ApiTag {
	names := make([]string, 0, len(folders))
	for _, folder := range folders {
		names = append(names, folder.Name)
	}
	description := ""
	if len(folders) > 0 {
		description = folders[len(folders)-1].Description
	}
	return ApiTag{Name: strings.Join(names, TagSeparator), Description: description}
}

// FolderNames splits a tag like "Products / Admin" into the folder names it stands for.
func FolderNames(tag string) []string {
	names := make([]string, 0)
	for _, name := range strings.Split(tag, strings.TrimSpace(TagSeparator)) {
		if trimmed := NormalizeItemName(name); trimmed != "" {
			names = append(names, trimmed)
		}
	}
	return names
}
