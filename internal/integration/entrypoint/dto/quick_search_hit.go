package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type QuickSearchHit struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Method string `json:"method"`
	Url    string `json:"url"`
	Path   string `json:"path"`
}

func NewQuickSearchHit(hit entity.QuickSearchHit) QuickSearchHit {
	return QuickSearchHit{Id: hit.ID, Name: hit.Name, Method: hit.Method, Url: hit.URL, Path: hit.Path}
}
