package model

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ClosedTab struct {
	ClosedAt time.Time `json:"closed_at"`
	Tab      Tab       `json:"tab"`
}

func NewClosedTab(closed entity.ClosedTab) ClosedTab {
	return ClosedTab{ClosedAt: closed.ClosedAt, Tab: NewTab(closed.Tab)}
}

func (c *ClosedTab) ToEntity() entity.ClosedTab {
	return entity.ClosedTab{ClosedAt: c.ClosedAt, Tab: c.Tab.ToEntity()}
}
