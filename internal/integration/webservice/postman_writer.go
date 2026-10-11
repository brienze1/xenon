package webservice

import (
	"context"
	"errors"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

var errPostmanWriterMissing = errors.New("the Postman writer is not implemented yet")

func (p *postmanWebservice) WriteCollection(ctx context.Context, root entity.ExportNode) ([]byte, error) {
	return nil, errs.InternalServerError(errPostmanWriterMissing, "EXP-06500", "the Postman collection could not be written")
}

func (p *postmanWebservice) WriteEnvironment(ctx context.Context, environment entity.Environment) ([]byte, error) {
	return nil, errs.InternalServerError(errPostmanWriterMissing, "EXP-07500", "the Postman environment could not be written")
}

func (p *postmanWebservice) WriteGlobals(ctx context.Context, workspace entity.Workspace) ([]byte, error) {
	return nil, errs.InternalServerError(errPostmanWriterMissing, "EXP-08500", "the Postman globals could not be written")
}
