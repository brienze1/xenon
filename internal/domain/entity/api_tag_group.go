package entity

import "slices"

// ApiTagGroup is a section of tags (OpenAPI's x-tagGroups): a top-level folder with its own tag and every nested folder's.
type ApiTagGroup struct {
	Name string
	Tags []string
}

// WithTagGroup files the tag of the folders under the group of their top-level folder, opening that group for a top-level one.
func WithTagGroup(groups []ApiTagGroup, folders []Collection, tag ApiTag) []ApiTagGroup {
	if len(folders) == 0 {
		return groups
	}
	groups = slices.Clone(groups)
	if len(folders) == 1 || len(groups) == 0 {
		return append(groups, ApiTagGroup{Name: folders[0].Name, Tags: []string{tag.Name}})
	}
	last := &groups[len(groups)-1]
	last.Tags = append(slices.Clone(last.Tags), tag.Name)
	return groups
}
