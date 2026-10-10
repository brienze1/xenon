package webservice

import (
	"context"
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/postman/dto"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
)

type postmanWebservice struct{}

func NewPostmanWebservice() adapter.PostmanWebservice {
	return &postmanWebservice{}
}

func (p *postmanWebservice) ReadCollection(ctx context.Context, document entity.ImportDocument) (*entity.ImportNode, []entity.ImportWarning, error) {
	logger.Debug(ctx, "Started", map[string]any{"path": document.Path, "bytes": len(document.Data)})

	var collection dto.Collection
	if err := json.Unmarshal(document.Data, &collection); err != nil {
		return nil, nil, errs.BadRequestError(document.Path+" is not a valid Postman collection: "+err.Error(), "IMP-33400")
	}
	conversion := dto.NewConversion(idutils.NewUlid)
	root := collection.ToEntity(conversion)

	logger.Debug(ctx, "Finished", map[string]any{"path": document.Path, "items": root.Counts().Requests, "warnings": len(conversion.Warnings())})
	return &root, conversion.Warnings(), nil
}
