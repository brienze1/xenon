package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type TreeSearch struct {
	MatchIds    []string `json:"match_ids"`
	AncestorIds []string `json:"ancestor_ids"`
}

func NewTreeSearch(search entity.TreeSearch) TreeSearch {
	return TreeSearch{MatchIds: append([]string{}, search.MatchIDs...), AncestorIds: append([]string{}, search.AncestorIDs...)}
}
