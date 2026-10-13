package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type QuickSearchHitList []QuickSearchHit

func NewQuickSearchHitList(hits []entity.QuickSearchHit) QuickSearchHitList {
	list := make(QuickSearchHitList, 0, len(hits))
	for _, hit := range hits {
		list = append(list, NewQuickSearchHit(hit))
	}
	return list
}
