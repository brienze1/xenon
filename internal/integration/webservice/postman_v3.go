package webservice

import (
	"context"
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/webservice/postmanv3/dto"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

func (p *postmanWebservice) ReadPostmanV3(ctx context.Context, input entity.ImportInput) (*entity.PostmanV3Source, error) {
	logger.Debug(ctx, "Started", input)

	source, err := readPostmanV3(input)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", source)
	return source, nil
}

func readPostmanV3(input entity.ImportInput) (*entity.PostmanV3Source, error) {
	source := entity.NewPostmanV3Source()
	switch {
	case input.IsPostmanWorkspace():
		return readPostmanV3Workspace(input)
	case input.IsPostmanRequest():
		request, err := readPostmanV3Request(input)
		if err != nil {
			return nil, err
		}
		source = source.WithRequest(*request)
	case input.IsPostmanEnvironment():
		environment, err := readPostmanV3Variables(input.Document(), (*dto.Environment).ToEnvironmentV21)
		if err != nil {
			return nil, err
		}
		source = source.WithEnvironment(*environment)
	case input.IsPostmanGlobals():
		globals, err := readPostmanV3Variables(input.Document(), (*dto.Environment).ToGlobalsV21)
		if err != nil {
			return nil, err
		}
		source = source.WithGlobals(*globals)
	default:
		collection, err := readPostmanV3Collection(input)
		if err != nil {
			return nil, err
		}
		source = source.WithCollection(*collection)
	}
	return &source, nil
}

func readPostmanV3Collection(input entity.ImportInput) (*entity.ImportCollection, error) {
	if !input.IsFolder() {
		var collection dto.Collection
		if err := dto.Decode(input.Document(), &collection); err != nil {
			return nil, invalidPostmanV3File(err)
		}
		if err := collection.ResolveItems(); err != nil {
			return nil, invalidPostmanV3File(dto.NewDecodeError(input.Document().Path, err))
		}
		return convertPostmanV3Collection(&collection, input.Document().Path)
	}
	collection, err := dto.NewFolderTree(input.Documents, input.Root)
	if err != nil {
		return nil, invalidPostmanV3File(err)
	}
	return convertPostmanV3Collection(collection, input.Name())
}

func readPostmanV3Request(input entity.ImportInput) (*entity.ImportCollection, error) {
	document := input.Document()
	collection, err := dto.NewTree(input.Documents, "", document.Path)
	if err != nil {
		return nil, invalidPostmanV3File(err)
	}
	imported, err := convertPostmanV3Collection(collection, document.Path)
	if err != nil {
		return nil, err
	}
	request, found := imported.FirstChild()
	if !found {
		return nil, invalidPostmanV3File(dto.NewDecodeError(document.Path, dto.NewUnknownKindError("")))
	}
	return &request, nil
}

func readPostmanV3Workspace(input entity.ImportInput) (*entity.PostmanV3Source, error) {
	workspace := dto.NewWorkspace(input.Documents, input.Entries())
	source := workspace.ToEntity()
	for _, prefix := range workspace.Collections {
		tree, err := dto.NewWorkspaceTree(input.Documents, prefix)
		if err != nil {
			return nil, invalidPostmanV3File(err)
		}
		collection, err := convertPostmanV3Collection(tree, prefix)
		if err != nil {
			return nil, err
		}
		source = source.WithCollection(*collection)
	}
	for _, document := range workspace.Environments {
		environment, err := readPostmanV3Variables(document, (*dto.Environment).ToEnvironmentV21)
		if err != nil {
			return nil, err
		}
		source = source.WithEnvironment(*environment)
	}
	for _, document := range workspace.Globals {
		globals, err := readPostmanV3Variables(document, (*dto.Environment).ToGlobalsV21)
		if err != nil {
			return nil, err
		}
		source = source.WithGlobals(*globals)
	}
	return &source, nil
}

func readPostmanV3Variables(document entity.ImportDocument, toPostmanV21 func(*dto.Environment, string, *dto.Conversion) map[string]any) (*entity.ImportVariables, error) {
	var environment dto.Environment
	if err := dto.Decode(document, &environment); err != nil {
		return nil, invalidPostmanV3File(err)
	}
	conversion := dto.NewConversion()
	data, err := json.Marshal(toPostmanV21(&environment, document.Path, conversion))
	if err != nil {
		return nil, invalidPostmanV3File(dto.NewDecodeError(document.Path, err))
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
		return nil, invalidPostmanV3File(dto.NewDecodeError(path, err))
	}
	data, err := json.Marshal(document)
	if err != nil {
		return nil, invalidPostmanV3File(dto.NewDecodeError(path, err))
	}
	imported, err := readPostmanCollection(entity.ImportDocument{Path: path, Data: data})
	if err != nil {
		return nil, err
	}
	warned := imported.WithItemWarnings(conversion.Warnings())
	return &warned, nil
}

func invalidPostmanV3File(reason error) error {
	return errs.BadRequestError(reason.Error(), "IMP-11400")
}
