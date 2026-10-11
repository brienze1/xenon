package webservice

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/webservice/postman/dto"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const (
	postmanExportedAtLayout = "2006-01-02T15:04:05.000Z"
	postmanExporterPrefix   = "Xenon/"
)

func (p *postmanWebservice) WriteCollection(ctx context.Context, root entity.ExportNode) ([]byte, error) {
	logger.Debug(ctx, "Started", map[string]any{"id": root.Container.ID, "name": root.Name()})

	data, err := dto.EncodeDocument(dto.NewCollection(root))
	if err != nil {
		return nil, errs.InternalServerError(err, "EXP-06500", "the Postman collection could not be written")
	}

	logger.Debug(ctx, "Finished", map[string]any{"id": root.Container.ID, "bytes": len(data)})
	return data, nil
}

func (p *postmanWebservice) WriteEnvironment(ctx context.Context, environment entity.Environment) ([]byte, error) {
	logger.Debug(ctx, "Started", environment)

	data, err := dto.EncodeDocument(dto.NewEnvironment(environment, exportedAt(), exportedUsing()))
	if err != nil {
		return nil, errs.InternalServerError(err, "EXP-07500", "the Postman environment could not be written")
	}

	logger.Debug(ctx, "Finished", map[string]any{"id": environment.ID, "bytes": len(data)})
	return data, nil
}

func (p *postmanWebservice) WriteGlobals(ctx context.Context, workspace entity.Workspace) ([]byte, error) {
	logger.Debug(ctx, "Started", workspace)

	data, err := dto.EncodeDocument(dto.NewGlobals(workspace, exportedAt(), exportedUsing()))
	if err != nil {
		return nil, errs.InternalServerError(err, "EXP-08500", "the Postman globals could not be written")
	}

	logger.Debug(ctx, "Finished", map[string]any{"id": workspace.ID, "bytes": len(data)})
	return data, nil
}

func exportedAt() string {
	return timeutils.Now().UTC().Format(postmanExportedAtLayout)
}

func exportedUsing() string {
	return postmanExporterPrefix + properties.Properties().Application.Version
}
