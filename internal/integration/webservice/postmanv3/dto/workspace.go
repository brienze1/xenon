package dto

import (
	"path"
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	postmanFolder   = "postman/"
	collectionsDir  = "collections"
	environmentsDir = "environments"
	globalsDir      = "globals"
)

var (
	environmentSuffixes = []string{".environment.yaml", ".environment.yml", ".environment.json"}
	globalsSuffixes     = []string{".globals.yaml", ".globals.yml", ".globals.json"}
	workspaceParts      = []string{collectionsDir, environmentsDir, globalsDir}
)

type Workspace struct {
	Collections  []string
	Environments []entity.ImportDocument
	Globals      []entity.ImportDocument
	Skipped      []entity.PostmanDumpPart
}

func (w Workspace) ToEntity() entity.PostmanV3Source {
	return entity.NewPostmanV3Source().WithSkipped(w.Skipped)
}

func NewWorkspace(documents []entity.ImportDocument, entries []string) Workspace {
	base := workspaceBase(entries)
	if base == "" && !slices.ContainsFunc(entries, isWorkspacePartEntry) {
		return Workspace{Collections: collectionPrefixes(entries, ""), Environments: []entity.ImportDocument{}, Globals: []entity.ImportDocument{}, Skipped: []entity.PostmanDumpPart{}}
	}
	workspace := Workspace{Collections: collectionPrefixes(entries, base+collectionsDir+"/"), Environments: []entity.ImportDocument{}, Globals: []entity.ImportDocument{}}
	for _, document := range documents {
		dir, file := path.Split(document.Path)
		switch {
		case dir == base+environmentsDir+"/" && hasSuffixFold(file, environmentSuffixes...):
			workspace.Environments = append(workspace.Environments, document)
		case dir == base+globalsDir+"/" && hasSuffixFold(file, globalsSuffixes...):
			workspace.Globals = append(workspace.Globals, document)
		}
	}
	slices.SortFunc(workspace.Environments, func(a, b entity.ImportDocument) int { return compareFold(a.Path, b.Path) })
	slices.SortFunc(workspace.Globals, func(a, b entity.ImportDocument) int { return compareFold(a.Path, b.Path) })
	workspace.Skipped = skippedParts(entries, base)
	return workspace
}

func workspaceBase(entries []string) string {
	if slices.ContainsFunc(entries, func(entry string) bool {
		return isWorkspacePartEntry(strings.TrimPrefix(entry, postmanFolder)) && strings.HasPrefix(entry, postmanFolder)
	}) {
		return postmanFolder
	}
	return ""
}

func isWorkspacePartEntry(entry string) bool {
	part, _, found := strings.Cut(entry, "/")
	return found && slices.Contains(workspaceParts, part)
}

func collectionPrefixes(entries []string, base string) []string {
	prefixes := []string{}
	for _, entry := range entries {
		relative, found := strings.CutPrefix(entry, base)
		name, rest, nested := strings.Cut(relative, "/")
		if !found || !nested || !isV3Entry(rest) {
			continue
		}
		if prefix := base + name + "/"; !slices.Contains(prefixes, prefix) {
			prefixes = append(prefixes, prefix)
		}
	}
	slices.SortFunc(prefixes, compareFold)
	return prefixes
}

func skippedParts(entries []string, base string) []entity.PostmanDumpPart {
	files := map[string]int{}
	for _, entry := range entries {
		relative, found := strings.CutPrefix(entry, base)
		part, _, nested := strings.Cut(relative, "/")
		if found && nested && !slices.Contains(workspaceParts, part) {
			files[part]++
		}
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	slices.Sort(names)
	parts := make([]entity.PostmanDumpPart, 0, len(names))
	for _, name := range names {
		parts = append(parts, entity.PostmanDumpPart{Name: name, Files: files[name]})
	}
	return parts
}

func collectionName(prefix string) string {
	return path.Base(strings.TrimSuffix(prefix, "/"))
}

func stemOf(documentPath string) string {
	base := path.Base(documentPath)
	for _, suffix := range append(slices.Clone(environmentSuffixes), globalsSuffixes...) {
		if hasSuffixFold(base, suffix) {
			return base[:len(base)-len(suffix)]
		}
	}
	return strings.TrimSuffix(base, path.Ext(base))
}

func isV3Entry(relative string) bool {
	return hasSuffixFold(relative, requestSuffix) || strings.HasSuffix(relative, resourcesDir+"/"+definitionFile)
}
