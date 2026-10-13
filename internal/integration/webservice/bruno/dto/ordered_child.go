package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type OrderedChild struct {
	build func() (entity.ImportNode, bool)
	name  string
	seq   int
}

func NewOrderedChild(name string, seq int, build func() (entity.ImportNode, bool)) OrderedChild {
	return OrderedChild{name: name, seq: seq, build: build}
}

func (o OrderedChild) isBefore(other OrderedChild) bool {
	if o.seq != other.seq {
		return o.seq < other.seq
	}
	return o.name < other.name
}
