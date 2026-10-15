package dto

import (
	"cmp"
	"math"
)

type Child struct {
	Folder      *Collection
	Request     *Request
	SkippedKind string
	Name        string
	SortName    string
	Order       float64
	HasOrder    bool
}

func NewFolderChild(folder *Collection, sortName string) Child {
	return newChild(Child{Folder: folder, Name: folder.Name, SortName: sortName}, folder.Order)
}

func NewRequestChild(request *Request, sortName string) Child {
	return newChild(Child{Request: request, Name: request.Name, SortName: sortName}, request.Order)
}

func NewSkippedChild(name, kind, sortName string, order *float64) Child {
	return newChild(Child{SkippedKind: kind, Name: name, SortName: sortName}, order)
}

func newChild(child Child, order *float64) Child {
	child.Order, child.HasOrder = math.MaxFloat64, order != nil
	if order != nil {
		child.Order = *order
	}
	return child
}

func compareChildren(a, b Child) int {
	if byOrder := cmp.Compare(a.Order, b.Order); byOrder != 0 {
		return byOrder
	}
	return compareFold(a.SortName, b.SortName)
}
