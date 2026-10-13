package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type Scripts struct {
	PreRequest       entity.BrunoScript
	PostResponse     entity.BrunoScript
	Tests            entity.BrunoScript
	PreRequestVars   []Pair
	PostResponseVars []Pair
	Assertions       []entity.BrunoAssertion
}

func (s Scripts) ToEntity(itemId, path, file string) entity.BrunoScripts {
	return entity.BrunoScripts{
		ItemID:           itemId,
		Path:             path,
		File:             file,
		PreRequest:       s.PreRequest,
		PostResponse:     s.PostResponse,
		Tests:            s.Tests,
		PreRequestVars:   variables(s.PreRequestVars),
		PostResponseVars: variables(s.PostResponseVars),
		Assertions:       s.Assertions,
	}
}
