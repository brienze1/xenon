package bruno

import (
	"context"
	"encoding/json"
	"math"
	"path"
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
	configNameKey      = "name"
)

var (
	yamlExtensions      = []string{".yml", ".yaml"}
	openCollectionFiles = []string{"opencollection.yml", "opencollection.yaml"}
)

type brunoWebservice struct{}

type directory struct {
	folder      *dto.Folder
	directories map[string]*directory
	name        string
	requests    []dto.Request
}

type brunoSource struct {
	root         *directory
	collection   dto.Collection
	environments []dto.Environment
}

func NewBrunoWebservice() adapter.BrunoWebservice {
	return &brunoWebservice{}
}

func (b *brunoWebservice) ReadCollection(ctx context.Context, input entity.ImportInput) (*entity.BrunoCollection, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": input.Probe.Name, "kind": input.Kind, "documents": len(input.Documents)})

	source := &brunoSource{root: newDirectory(""), collection: dto.Collection{Name: input.Probe.Name}, environments: []dto.Environment{}}
	documents := append([]entity.ImportDocument{}, input.Documents...)
	sort.Slice(documents, func(left, right int) bool { return documents[left].Path < documents[right].Path })
	for _, document := range documents {
		if err := source.add(document, input.Kind == importkind.BrunoYaml); err != nil {
			return nil, err
		}
	}
	collection := source.toEntity(idutils.NewUlid)

	logger.Debug(ctx, "Finished", map[string]any{"root": input.Probe.Name, "requests": collection.Root.Counts().Requests, "environments": len(collection.Environments), "warnings": len(collection.Warnings)})
	return &collection, nil
}

func (s *brunoSource) add(document entity.ImportDocument, isYaml bool) error {
	if isYaml {
		return s.addYaml(document)
	}
	return s.addBru(document)
}

func (s *brunoSource) addBru(document entity.ImportDocument) error {
	dir, base := path.Split(document.Path)
	if document.Path == brunoConfigFile {
		return s.addConfig(document)
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
		s.collection = dto.NewCollectionFromBru(*file, s.collection.Name, s.collection.Config)
	case dir == environmentsFolder:
		s.environments = append(s.environments, dto.NewEnvironmentFromBru(*file, baseName(base)))
	case baseName(base) == folderBaseName:
		folder := dto.NewFolderFromBru(*file, path.Base(dir))
		s.directory(dir).folder = &folder
	default:
		s.directory(dir).requests = append(s.directory(dir).requests, dto.NewRequestFromBru(*file))
	}
	return nil
}

func (s *brunoSource) addConfig(document entity.ImportDocument) error {
	config := map[string]any{}
	if err := json.Unmarshal(document.Data, &config); err != nil {
		return errs.BadRequestError(document.Path+" is not valid JSON: "+err.Error(), "IMP-37400")
	}
	if name, isText := config[configNameKey].(string); isText && strings.TrimSpace(name) != "" {
		s.collection.Name = name
	}
	delete(config, configNameKey)
	s.collection.Config = config
	return nil
}

func (s *brunoSource) addYaml(document entity.ImportDocument) error {
	dir, base := path.Split(document.Path)
	if !contains(yamlExtensions, strings.ToLower(path.Ext(base))) {
		return nil
	}
	switch {
	case contains(openCollectionFiles, strings.ToLower(document.Path)):
		var collection dto.YamlCollection
		return decodeYaml(document, &collection, func() { s.collection = collection.ToCollection(document.Path) })
	case dir == environmentsFolder:
		var environment dto.YamlEnvironment
		return decodeYaml(document, &environment, func() { s.environments = append(s.environments, environment.ToEnvironment(baseName(base))) })
	case baseName(base) == folderBaseName:
		var folder dto.YamlFolder
		return decodeYaml(document, &folder, func() {
			converted := folder.ToFolder(document.Path, path.Base(dir))
			s.directory(dir).folder = &converted
		})
	default:
		var request dto.YamlRequest
		return decodeYaml(document, &request, func() {
			s.directory(dir).requests = append(s.directory(dir).requests, request.ToRequest(document.Path))
		})
	}
}

func (s *brunoSource) directory(dir string) *directory {
	current := s.root
	for _, segment := range strings.Split(strings.Trim(dir, "/"), "/") {
		if segment == "" {
			continue
		}
		if current.directories[segment] == nil {
			current.directories[segment] = newDirectory(segment)
		}
		current = current.directories[segment]
	}
	return current
}

func (s *brunoSource) toEntity(newId func() string) entity.BrunoCollection {
	conversion := dto.NewConversion(newId)
	collection := s.collection.ToEntity(conversion)
	root := entity.NewContainerImportNode(collection, s.root.children(conversion, collection, collection.Name))
	environments := make([]entity.ImportVariables, 0, len(s.environments))
	for index := range s.environments {
		environments = append(environments, s.environments[index].ToEntity(newId))
	}
	return entity.BrunoCollection{Root: root, Environments: environments, Scripts: conversion.Scripts(), Warnings: conversion.Warnings()}
}

type orderedChild struct {
	build func() (entity.ImportNode, bool)
	name  string
	seq   int
}

func (d *directory) children(conversion *dto.Conversion, parent entity.Collection, parentPath string) []entity.ImportNode {
	ordered := make([]orderedChild, 0, len(d.directories)+len(d.requests))
	for _, sub := range d.directories {
		ordered = append(ordered, orderedChild{name: sub.name, seq: sub.seq(), build: func() (entity.ImportNode, bool) { return sub.toEntity(conversion, parentPath), true }})
	}
	for index := range d.requests {
		request := &d.requests[index]
		if strings.TrimSpace(request.Name) == "" {
			request.Name = baseName(path.Base(request.File))
		}
		ordered = append(ordered, orderedChild{name: path.Base(request.File), seq: seqOf(request.Seq), build: func() (entity.ImportNode, bool) { return request.ToEntity(conversion, parent, parentPath) }})
	}
	sort.SliceStable(ordered, func(left, right int) bool {
		if ordered[left].seq != ordered[right].seq {
			return ordered[left].seq < ordered[right].seq
		}
		return ordered[left].name < ordered[right].name
	})
	nodes := make([]entity.ImportNode, 0, len(ordered))
	for _, child := range ordered {
		if node, kept := child.build(); kept {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (d *directory) toEntity(conversion *dto.Conversion, parentPath string) entity.ImportNode {
	source := dto.Folder{Name: d.name}
	if d.folder != nil {
		source = *d.folder
	}
	path := entity.JoinImportPath(parentPath, source.Name)
	folder := source.ToEntity(conversion, path)
	return entity.NewContainerImportNode(folder, d.children(conversion, folder, path))
}

func (d *directory) seq() int {
	if d.folder == nil {
		return math.MaxInt
	}
	return seqOf(d.folder.Seq)
}

func newDirectory(name string) *directory {
	return &directory{name: name, directories: map[string]*directory{}, requests: []dto.Request{}}
}

func decodeYaml(document entity.ImportDocument, target any, apply func()) error {
	if err := yaml.Unmarshal(document.Data, target); err != nil {
		return errs.BadRequestError(document.Path+" is not valid Bruno YAML: "+err.Error(), "IMP-37400")
	}
	apply()
	return nil
}

func seqOf(seq *int) int {
	if seq == nil {
		return math.MaxInt
	}
	return *seq
}

func baseName(file string) string {
	return strings.TrimSuffix(file, path.Ext(file))
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
