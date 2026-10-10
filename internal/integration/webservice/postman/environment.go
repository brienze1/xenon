package postman

import (
	"context"
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/webservice/postman/dto"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
)

func (p *postmanWebservice) ReadEnvironment(ctx context.Context, document entity.ImportDocument) (*entity.ImportVariables, error) {
	logger.Debug(ctx, "Started", map[string]any{"path": document.Path, "bytes": len(document.Data)})

	var environment dto.Environment
	if err := json.Unmarshal(document.Data, &environment); err != nil {
		return nil, errs.BadRequestError(document.Path+" is not a valid Postman environment: "+err.Error(), "IMP-34400")
	}
	variables := environment.ToEntity(idutils.NewUlid)

	logger.Debug(ctx, "Finished", map[string]any{"path": document.Path, "result": variables})
	return &variables, nil
}
