package dto

import (
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	resourcesDir   = ".resources"
	definitionFile = "definition.yaml"
	requestSuffix  = ".request.yaml"
	exampleSuffix  = ".example.yaml"
	examplesDir    = "examples"
)

type tree struct {
	folders  map[string]*TreeFolder
	examples map[string][]TreeDocument
}

func NewTree(documents []entity.ImportDocument, prefix, name string) (*Collection, error) {
	built := tree{folders: map[string]*TreeFolder{"": NewTreeFolder()}, examples: map[string][]TreeDocument{}}
	for _, document := range documents {
		relative, found := strings.CutPrefix(document.Path, prefix)
		if found {
			built.add(NewTreeDocument(document, relative))
		}
	}
	return built.collection("", name)
}

func NewFolderTree(documents []entity.ImportDocument, root string) (*Collection, error) {
	return NewTree(documents, "", filepath.Base(root))
}

func NewWorkspaceTree(documents []entity.ImportDocument, prefix string) (*Collection, error) {
	return NewTree(documents, prefix, collectionName(prefix))
}

func (t tree) add(document TreeDocument) {
	dir, base := path.Split(document.Relative)
	dir = strings.TrimSuffix(dir, "/")
	switch {
	case slices.Contains(strings.Split(dir, "/"), resourcesDir):
		if base == definitionFile && path.Base(dir) == resourcesDir {
			t.folder(parentDir(dir)).Definition = &document
		} else if hasSuffixFold(base, exampleSuffix) {
			t.examples[dir] = append(t.examples[dir], document)
		}
	case hasSuffixFold(base, requestSuffix):
		folder := t.folder(dir)
		folder.Requests = append(folder.Requests, document)
	}
}

func (t tree) folder(dir string) *TreeFolder {
	if folder, found := t.folders[dir]; found {
		return folder
	}
	folder := NewTreeFolder()
	t.folders[dir] = folder
	parent := t.folder(parentDir(dir))
	parent.Subfolders = append(parent.Subfolders, dir)
	return folder
}

func (t tree) collection(dir, name string) (*Collection, error) {
	folder := t.folders[dir]
	collection, err := NewFolderCollection(folder.Definition, name)
	if err != nil {
		return nil, err
	}
	for _, sub := range folder.Subfolders {
		child, err := t.collection(sub, path.Base(sub))
		if err != nil {
			return nil, err
		}
		collection.Children = append(collection.Children, NewFolderChild(child, path.Base(sub)))
	}
	for _, document := range folder.Requests {
		child, err := t.request(dir, document)
		if err != nil {
			return nil, err
		}
		collection.Children = append(collection.Children, child)
	}
	slices.SortStableFunc(collection.Children, compareChildren)
	return collection, nil
}

func (t tree) request(dir string, document TreeDocument) (Child, error) {
	base := path.Base(document.Relative)
	stem := base[:len(base)-len(requestSuffix)]
	var request Request
	if err := Decode(document.Document, &request); err != nil {
		return Child{}, err
	}
	if request.Name == "" {
		request.Name = stem
	}
	switch {
	case IsSupportedRequestKind(request.Kind):
		examples, err := t.requestExamples(dir, stem, request.examplesPath())
		return NewRequestChild(request.withExamples(examples), base), err
	case IsRequestKind(request.Kind):
		return NewSkippedChild(request.Name, request.Kind, base, request.Order), nil
	default:
		return Child{}, NewDecodeError(document.Document.Path, NewUnknownKindError(request.Kind))
	}
}

func (t tree) requestExamples(dir, stem, examplesPath string) ([]Example, error) {
	location := path.Join(dir, resourcesDir, stem+".resources", examplesDir)
	if examplesPath != "" {
		location = path.Join(dir, strings.TrimSuffix(examplesPath, "/"))
	}
	documents := t.examples[location]
	examples := make([]Example, 0, len(documents))
	for _, document := range documents {
		var example Example
		if err := Decode(document.Document, &example); err != nil {
			return nil, err
		}
		if example.Name == "" {
			base := path.Base(document.Relative)
			example.Name = base[:len(base)-len(exampleSuffix)]
		}
		examples = append(examples, example)
	}
	slices.SortStableFunc(examples, func(a, b Example) int { return compareChildren(a.sortKey(), b.sortKey()) })
	return examples, nil
}

func parentDir(dir string) string {
	parent := path.Dir(dir)
	if parent == "." {
		return ""
	}
	return parent
}
