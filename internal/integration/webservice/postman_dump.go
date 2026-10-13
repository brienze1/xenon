package webservice

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/webservice/postman/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

func (p *postmanWebservice) ReadDump(ctx context.Context, input entity.ImportInput) (*entity.PostmanDump, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": input.Probe.Name, "documents": len(input.Documents)})

	source := dto.NewDump(input.Documents, input.Probe.Entries)
	dump := source.ToEntity()

	logger.Debug(ctx, "Finished", map[string]any{"root": input.Probe.Name, "collections": len(dump.Collections), "environments": len(dump.Environments), "skipped": len(dump.Skipped)})
	return &dump, nil
}
