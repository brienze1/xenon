package persistence

import (
	"context"
	"path"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

func (t *treePersistence) RenderItem(ctx context.Context, kind enums.ItemKind, filePath string, content []byte) entity.ConflictVersion {
	logger.Debug(ctx, "Started", map[string]any{"kind": kind, "path": filePath, "bytes": len(content)})

	version := entity.ConflictVersion{Text: string(content)}
	switch kind {
	case itemkind.Request:
		var stored model.Request
		if decodeDocument(path.Base(filePath), content, itemkind.Request, &stored.Xenon, &stored.Kind, &stored) == "" {
			request := stored.ToStoredEntity(stored.Id, revisionOf(content))
			version.Request = &request
		}
	case itemkind.Environment:
		var stored model.Environment
		if decodeDocument(path.Base(filePath), content, itemkind.Environment, &stored.Xenon, &stored.Kind, &stored) == "" {
			environment := stored.ToEntity(path.Base(filePath))
			version.Environment = &environment
		}
	}

	logger.Debug(ctx, "Finished", map[string]any{"kind": kind, "path": filePath, "request": version.Request != nil, "environment": version.Environment != nil})
	return version
}

func (t *treePersistence) RenderNode(ctx context.Context, node *entity.DocumentNode) (*string, error) {
	logger.Debug(ctx, "Started")

	if node == nil {
		logger.Debug(ctx, "Finished", map[string]any{"present": false})
		return nil, nil
	}
	text := node.Value
	if !node.IsScalar() {
		data, err := yamlnode.Encode(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{model.DocumentYaml(node)}})
		if err != nil {
			return nil, errs.InternalServerError(err, "GIT-37500", "the conflicting field could not be shown")
		}
		text = strings.TrimSuffix(string(data), "\n")
	}

	logger.Debug(ctx, "Finished", map[string]any{"present": true})
	return &text, nil
}
