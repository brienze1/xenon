package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func ResolveExample(request entity.Request, exampleId string) (*entity.SavedExample, error) {
	example, found := request.Example(exampleId)
	if !found {
		return nil, errs.NotFoundError("the example does not exist", "EXM-01404")
	}
	return &example, nil
}
