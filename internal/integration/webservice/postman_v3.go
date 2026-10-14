package webservice

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/importkind"
	"github.com/brienze1/xenon/internal/integration/webservice/postmanv3/dto"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	environmentScope = "environment"
	globalsScope     = "globals"
)

func (p *postmanWebservice) ReadPostmanV3(ctx context.Context, input entity.ImportInput) (*entity.PostmanV3Source, error) {
	logger.Debug(ctx, "Started", input)

	source, err := readPostmanV3(input)
	if err != nil {
		return nil, postmanV3Error(err)
	}

	logger.Debug(ctx, "Finished", source)
	return source, nil
}

func readPostmanV3(input entity.ImportInput) (*entity.PostmanV3Source, error) {
	source := entity.NewPostmanV3Source()
	switch input.Kind {
	case importkind.PostmanWorkspace:
		return readPostmanV3Workspace(input, source)
	case importkind.PostmanRequest:
		request, err := readPostmanV3Request(input)
		source.Request = request
		return &source, err
	case importkind.PostmanEnvironment:
		environment, err := readPostmanV3Variables(input.Document(), environmentScope)
		if err != nil {
			return nil, err
		}
		source.Environments = append(source.Environments, *environment)
		return &source, nil
	case importkind.PostmanGlobals:
		globals, err := readPostmanV3Variables(input.Document(), globalsScope)
		source.Globals = globals
		return &source, err
	default:
		collection, err := readPostmanV3Collection(input)
		if err != nil {
			return nil, err
		}
		source.Collections = append(source.Collections, *collection)
		return &source, nil
	}
}

func readPostmanV3Collection(input entity.ImportInput) (*entity.ImportCollection, error) {
	if !input.Probe.IsFolder {
		var collection dto.Collection
		if err := dto.Decode(input.Document(), &collection); err != nil {
			return nil, err
		}
		if err := collection.ResolveItems(); err != nil {
			return nil, dto.NewDecodeError(input.Document().Path, err)
		}
		return convertPostmanV3Collection(&collection, input.Document().Path)
	}
	collection, err := dto.NewTree(input.Documents, "", filepath.Base(input.Root))
	if err != nil {
		return nil, err
	}
	return convertPostmanV3Collection(collection, input.Probe.Name)
}

func readPostmanV3Request(input entity.ImportInput) (*entity.ImportCollection, error) {
	document := input.Document()
	collection, err := dto.NewTree(input.Documents, "", document.Path)
	if err != nil {
		return nil, err
	}
	imported, err := convertPostmanV3Collection(collection, document.Path)
	if err != nil {
		return nil, err
	}
	if len(imported.Root.Children) == 0 {
		return nil, dto.NewDecodeError(document.Path, dto.UnknownKindError{Kind: ""})
	}
	request := entity.NewImportCollection(imported.Root.Children[0], imported.Warnings, imported.Secrets)
	return &request, nil
}

func readPostmanV3Workspace(input entity.ImportInput, source entity.PostmanV3Source) (*entity.PostmanV3Source, error) {
	workspace := dto.NewWorkspace(input.Documents, input.Probe.Entries)
	for _, prefix := range workspace.Collections {
		tree, err := dto.NewTree(input.Documents, prefix, dto.CollectionName(prefix))
		if err != nil {
			return nil, err
		}
		collection, err := convertPostmanV3Collection(tree, prefix)
		if err != nil {
			return nil, err
		}
		source.Collections = append(source.Collections, *collection)
	}
	for _, document := range workspace.Environments {
		environment, err := readPostmanV3Variables(document, environmentScope)
		if err != nil {
			return nil, err
		}
		source.Environments = append(source.Environments, *environment)
	}
	for _, document := range workspace.Globals {
		globals, err := readPostmanV3Variables(document, globalsScope)
		if err != nil {
			return nil, err
		}
		source.Globals = globals
	}
	source.Skipped = workspace.Skipped
	return &source, nil
}

func readPostmanV3Variables(document entity.ImportDocument, scope string) (*entity.ImportVariables, error) {
	var environment dto.Environment
	if err := dto.Decode(document, &environment); err != nil {
		return nil, err
	}
	conversion := dto.NewConversion()
	data, err := json.Marshal(environment.ToPostmanV21(scope, dto.StemOf(document.Path), conversion))
	if err != nil {
		return nil, dto.NewDecodeError(document.Path, err)
	}
	variables, err := readPostmanEnvironment(entity.ImportDocument{Path: document.Path, Data: data})
	if err != nil {
		return nil, err
	}
	warned := variables.WithWarnings(conversion.Warnings())
	return &warned, nil
}

func convertPostmanV3Collection(collection *dto.Collection, path string) (*entity.ImportCollection, error) {
	conversion := dto.NewConversion()
	document, err := collection.ToPostmanV21(conversion)
	if err != nil {
		return nil, dto.NewDecodeError(path, err)
	}
	data, err := json.Marshal(document)
	if err != nil {
		return nil, dto.NewDecodeError(path, err)
	}
	imported, err := readPostmanCollection(entity.ImportDocument{Path: path, Data: data})
	if err != nil {
		return nil, err
	}
	ids := imported.Root.IDsByPath("")
	for _, warning := range conversion.Warnings() {
		imported.Warnings = append(imported.Warnings, warning.OnItem(ids[warning.Path]))
	}
	return imported, nil
}

func postmanV3Error(err error) error {
	var decodeError dto.DecodeError
	if errors.As(err, &decodeError) {
		return errs.BadRequestError(decodeError.Error(), "IMP-11400")
	}
	return err
}
