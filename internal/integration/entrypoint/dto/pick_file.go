package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type PickFile struct {
	Title   string       `json:"title" validate:"required,not_blank"`
	Filters []FileFilter `json:"filters" validate:"dive"`
}

func (p *PickFile) FileFilters() []entity.FileFilter {
	filters := make([]entity.FileFilter, 0, len(p.Filters))
	for index := range p.Filters {
		filters = append(filters, p.Filters[index].ToEntity())
	}
	return filters
}
