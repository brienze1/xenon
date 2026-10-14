package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/historysource"
)

const historySlackDivisor = 5

var DefaultHistoryCaps = HistoryCaps{User: DefaultHistoryCap, AI: DefaultHistoryCap}

type HistoryCaps struct {
	User int
	AI   int
}

func IsValidHistoryCap(limit int) bool {
	return limit >= MinHistoryCap && limit <= MaxHistoryCap
}

func (h HistoryCaps) OrDefault() HistoryCaps {
	if !IsValidHistoryCap(h.User) {
		h.User = DefaultHistoryCap
	}
	if !IsValidHistoryCap(h.AI) {
		h.AI = DefaultHistoryCap
	}
	return h
}

func (h HistoryCaps) For(source enums.HistorySource) int {
	if source == historysource.AI {
		return h.AI
	}
	return h.User
}

func (h HistoryCaps) compactionThreshold(source enums.HistorySource) int {
	limit := h.For(source)
	return limit + limit/historySlackDivisor
}
