package model

import "github.com/brienze1/xenon/internal/domain/entity"

type BodyGraphQl struct {
	Query     string `yaml:"query" json:"query"`
	Variables string `yaml:"variables" json:"variables"`
}

func NewBodyGraphQl(graphQl entity.BodyGraphQl) BodyGraphQl {
	return BodyGraphQl{Query: graphQl.Query, Variables: graphQl.Variables}
}

func (b *BodyGraphQl) ToEntity() entity.BodyGraphQl {
	return entity.BodyGraphQl{Query: b.Query, Variables: b.Variables}
}
