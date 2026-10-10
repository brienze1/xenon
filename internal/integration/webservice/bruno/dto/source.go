package dto

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const configNameKey = "name"

type Source struct {
	Root         *Directory
	Collection   Collection
	Environments []Environment
}

func NewSource(name string) *Source {
	return &Source{Root: NewDirectory(""), Collection: NewCollection(name), Environments: []Environment{}}
}

func (s *Source) ApplyConfig(config map[string]any) {
	if name, isText := config[configNameKey].(string); isText && strings.TrimSpace(name) != "" {
		s.Collection.Name = name
	}
	delete(config, configNameKey)
	s.Collection.Config = config
}

func (s *Source) SetBruCollection(file BruFile) {
	s.Collection = NewCollectionFromBru(file, s.Collection.Name, s.Collection.Config)
}

func (s *Source) SetCollection(collection Collection) {
	s.Collection = collection
}

func (s *Source) AddEnvironment(environment Environment) {
	s.Environments = append(s.Environments, environment)
}

func (s *Source) SetFolder(dir string, folder Folder) {
	s.directory(dir).folder = &folder
}

func (s *Source) AddRequest(dir string, request Request) {
	target := s.directory(dir)
	target.requests = append(target.requests, request)
}

func (s *Source) ToEntity(newId func() string) entity.BrunoCollection {
	conversion := NewConversion(newId)
	collection := s.Collection.ToEntity(conversion)
	root := entity.NewContainerImportNode(collection, s.Root.children(conversion, collection, collection.Name))
	environments := make([]entity.ImportVariables, 0, len(s.Environments))
	for index := range s.Environments {
		environments = append(environments, s.Environments[index].ToEntity(newId))
	}
	return entity.BrunoCollection{Root: root, Environments: environments, Scripts: conversion.Scripts(), Warnings: conversion.Warnings()}
}

func (s *Source) directory(dir string) *Directory {
	current := s.Root
	for _, segment := range strings.Split(strings.Trim(dir, "/"), "/") {
		if segment == "" {
			continue
		}
		if current.directories[segment] == nil {
			current.directories[segment] = NewDirectory(segment)
		}
		current = current.directories[segment]
	}
	return current
}
