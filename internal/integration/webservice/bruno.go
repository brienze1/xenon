package webservice

import (
	"context"
	"encoding/json"
	"path"
	"slices"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/importkind"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/bruno/dto"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	brunoConfigFile    = "bruno.json"
	collectionBruFile  = "collection.bru"
	folderBaseName     = "folder"
	environmentsFolder = "environments/"
	bruExtension       = ".bru"
)

var (
	yamlExtensions      = []string{".yml", ".yaml"}
	openCollectionFiles = []string{"opencollection.yml", "opencollection.yaml"}
)

type brunoWebservice struct{}

func NewBrunoWebservice() adapter.BrunoWebservice {
	return &brunoWebservice{}
}

func (b *brunoWebservice) ReadCollection(ctx context.Context, input entity.ImportInput) (*entity.BrunoCollection, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": input.Probe.Name, "kind": input.Kind, "documents": len(input.Documents)})

	source := dto.NewSource(input.Probe.Name)
	documents := slices.Clone(input.Documents)
	sort.Slice(documents, func(left, right int) bool { return documents[left].Path < documents[right].Path })
	for _, document := range documents {
		if err := addBrunoDocument(source, document, input.Kind == importkind.BrunoYaml); err != nil {
			return nil, err
		}
	}
	collection := source.ToEntity(idutils.NewUlid)

	logger.Debug(ctx, "Finished", map[string]any{"root": input.Probe.Name, "requests": collection.Root.Counts().Requests, "environments": len(collection.Environments), "warnings": len(collection.Warnings)})
	return &collection, nil
}

func addBrunoDocument(source *dto.Source, document entity.ImportDocument, isYaml bool) error {
	if isYaml {
		return addYamlDocument(source, document)
	}
	return addBruDocument(source, document)
}

func addBruDocument(source *dto.Source, document entity.ImportDocument) error {
	dir, base := path.Split(document.Path)
	if document.Path == brunoConfigFile {
		return addBrunoConfig(source, document)
	}
	if !strings.EqualFold(path.Ext(base), bruExtension) {
		return nil
	}
	file, err := parseBru(document.Path, document.Data)
	if err != nil {
		return err
	}
	switch {
	case document.Path == collectionBruFile:
		source.SetBruCollection(*file)
	case dir == environmentsFolder:
		source.AddEnvironment(dto.NewEnvironmentFromBru(*file, baseName(base)))
	case baseName(base) == folderBaseName:
		source.SetFolder(dir, dto.NewFolderFromBru(*file, path.Base(dir)))
	default:
		source.AddRequest(dir, dto.NewRequestFromBru(*file))
	}
	return nil
}

func addBrunoConfig(source *dto.Source, document entity.ImportDocument) error {
	config := map[string]any{}
	if err := json.Unmarshal(document.Data, &config); err != nil {
		return errs.BadRequestError(document.Path+" is not valid JSON: "+err.Error(), "IMP-38400")
	}
	source.ApplyConfig(config)
	return nil
}

func addYamlDocument(source *dto.Source, document entity.ImportDocument) error {
	dir, base := path.Split(document.Path)
	if !slices.Contains(yamlExtensions, strings.ToLower(path.Ext(base))) {
		return nil
	}
	switch {
	case slices.Contains(openCollectionFiles, strings.ToLower(document.Path)):
		var collection dto.YamlCollection
		return decodeYaml(document, &collection, func() { source.SetCollection(dto.NewCollectionFromYaml(collection, document.Path)) })
	case dir == environmentsFolder:
		var environment dto.YamlEnvironment
		return decodeYaml(document, &environment, func() { source.AddEnvironment(dto.NewEnvironmentFromYaml(environment, baseName(base))) })
	case baseName(base) == folderBaseName:
		var folder dto.YamlFolder
		return decodeYaml(document, &folder, func() { source.SetFolder(dir, dto.NewFolderFromYaml(folder, document.Path, path.Base(dir))) })
	default:
		var request dto.YamlRequest
		return decodeYaml(document, &request, func() { source.AddRequest(dir, dto.NewRequestFromYaml(request, document.Path)) })
	}
}

func decodeYaml(document entity.ImportDocument, target any, apply func()) error {
	if err := yaml.Unmarshal(document.Data, target); err != nil {
		return errs.BadRequestError(document.Path+" is not valid Bruno YAML: "+err.Error(), "IMP-37400")
	}
	apply()
	return nil
}

func baseName(file string) string {
	return strings.TrimSuffix(file, path.Ext(file))
}
