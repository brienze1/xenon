package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AutoHeaderList []AutoHeader

func NewAutoHeaderList(headers []entity.AutoHeader) AutoHeaderList {
	list := make(AutoHeaderList, 0, len(headers))
	for _, header := range headers {
		list = append(list, NewAutoHeader(header))
	}
	return list
}

func (a AutoHeaderList) LogValue() any {
	return map[string]any{"headers": len(a)}
}
