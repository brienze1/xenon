package dto

import (
	"math"
	"path"
	"sort"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type Directory struct {
	folder      *Folder
	directories map[string]*Directory
	name        string
	requests    []Request
}

func NewDirectory(name string) *Directory {
	return &Directory{name: name, directories: map[string]*Directory{}, requests: []Request{}}
}

func (d *Directory) children(conversion *Conversion, parent entity.Collection, parentPath string) []entity.ImportNode {
	ordered := make([]OrderedChild, 0, len(d.directories)+len(d.requests))
	for _, sub := range d.directories {
		ordered = append(ordered, NewOrderedChild(sub.name, sub.seq(), func() (entity.ImportNode, bool) { return sub.toEntity(conversion, parentPath), true }))
	}
	for index := range d.requests {
		request := &d.requests[index]
		ordered = append(ordered, NewOrderedChild(path.Base(request.File), seqOf(request.Seq), func() (entity.ImportNode, bool) { return request.ToEntity(conversion, parent, parentPath) }))
	}
	sort.SliceStable(ordered, func(left, right int) bool { return ordered[left].isBefore(ordered[right]) })
	nodes := make([]entity.ImportNode, 0, len(ordered))
	for _, child := range ordered {
		if node, kept := child.build(); kept {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (d *Directory) toEntity(conversion *Conversion, parentPath string) entity.ImportNode {
	source := NewFolder(d.name)
	if d.folder != nil {
		source = *d.folder
	}
	folderPath := entity.JoinImportPath(parentPath, source.Name)
	folder := source.ToEntity(conversion, folderPath)
	return entity.NewContainerImportNode(folder, d.children(conversion, folder, folderPath))
}

func (d *Directory) seq() int {
	if d.folder == nil {
		return math.MaxInt
	}
	return seqOf(d.folder.Seq)
}

func seqOf(seq *int) int {
	if seq == nil {
		return math.MaxInt
	}
	return *seq
}
