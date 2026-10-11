package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type QuickSearch struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Query   string `json:"query"`
	Limit   int    `json:"limit" validate:"min=0"`
}

type QuickSearchHit struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Method string `json:"method"`
	Url    string `json:"url"`
	Path   string `json:"path"`
}

type QuickSearchHitList []QuickSearchHit

func NewQuickSearchHitList(hits []entity.QuickSearchHit) QuickSearchHitList {
	list := make(QuickSearchHitList, 0, len(hits))
	for _, hit := range hits {
		list = append(list, QuickSearchHit{Id: hit.ID, Name: hit.Name, Method: hit.Method, Url: hit.URL, Path: hit.Path})
	}
	return list
}
