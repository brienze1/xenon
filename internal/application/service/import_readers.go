package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/importitemkind"
	"github.com/brienze1/xenon/internal/domain/enums/importkind"
	"github.com/brienze1/xenon/pkg/idutils"
)

type importReaders struct {
	readPostmanCollection  usecase.ReadPostmanCollection
	readPostmanEnvironment usecase.ReadPostmanEnvironment
	readPostmanDump        usecase.ReadPostmanDump
	readBrunoCollection    usecase.ReadBrunoCollection
	readCurlCommand        usecase.ReadCurlCommand
	readXenonExport        usecase.ReadXenonExport
}

func (i importReaders) read(ctx context.Context, input entity.ImportInput, read *importRead) error {
	switch input.Kind {
	case importkind.PostmanCollection:
		return i.collection(ctx, input.Document(), read)
	case importkind.PostmanEnvironment:
		return i.variables(ctx, input.Document(), importitemkind.Environment, read)
	case importkind.PostmanGlobals:
		return i.variables(ctx, input.Document(), importitemkind.Globals, read)
	case importkind.PostmanDump:
		return i.dump(ctx, input, read)
	case importkind.BrunoBru, importkind.BrunoYaml:
		return i.bruno(ctx, input, read)
	case importkind.XenonWorkspaceZip, importkind.XenonWorkspaceFolder, importkind.XenonBundle, importkind.XenonZip, importkind.XenonRequest, importkind.XenonEnvironment:
		return i.xenon(ctx, input, read)
	default:
		return i.curl(ctx, input, read)
	}
}

func (i importReaders) collection(ctx context.Context, document entity.ImportDocument, read *importRead) error {
	root, warnings, err := i.readPostmanCollection.ReadCollection(ctx, document)
	if err != nil {
		return err
	}
	read.addTree(importitemkind.Collection, *root, warnings)
	return nil
}

func (i importReaders) bruno(ctx context.Context, input entity.ImportInput, read *importRead) error {
	collection, err := i.readBrunoCollection.ReadCollection(ctx, input)
	if err != nil {
		return err
	}
	scripts, warnings := translateBrunoScripts(collection.Scripts)
	read.addTree(importitemkind.Collection, collection.Root.WithScripts(scripts), append(collection.Warnings, warnings...))
	for _, environment := range collection.Environments {
		read.items = append(read.items, entity.NewVariablesImportItem(read.nextKey(), importitemkind.Environment, environment))
	}
	return nil
}

func (i importReaders) curl(ctx context.Context, input entity.ImportInput, read *importRead) error {
	parsed, err := i.readCurlCommand.ReadCommand(ctx, input.Probe.Text)
	if err != nil {
		return err
	}
	request := usecase.DeriveParamsFromUrl(parsed.Request)
	request = request.Identified(idutils.NewUlid(), usecase.DeriveCurlRequestName(request))
	read.addTree(importitemkind.Request, entity.NewRequestImportNode(request), []entity.ImportWarning{})
	return nil
}

func (i importReaders) variables(ctx context.Context, document entity.ImportDocument, kind enums.ImportItemKind, read *importRead) error {
	variables, err := i.readPostmanEnvironment.ReadEnvironment(ctx, document)
	if err != nil {
		return err
	}
	read.items = append(read.items, entity.NewVariablesImportItem(read.nextKey(), kind, *variables))
	return nil
}

func (i importReaders) dump(ctx context.Context, input entity.ImportInput, read *importRead) error {
	dump, err := i.readPostmanDump.ReadDump(ctx, input)
	if err != nil {
		return err
	}
	for _, document := range dump.Collections {
		if err := i.collection(ctx, document, read); err != nil {
			return err
		}
	}
	for _, document := range dump.Environments {
		if err := i.variables(ctx, document, importitemkind.Environment, read); err != nil {
			return err
		}
	}
	read.warnings = append(read.warnings, dump.SkippedWarnings()...)
	return nil
}

func (i importReaders) xenon(ctx context.Context, input entity.ImportInput, read *importRead) error {
	exported, err := i.readXenonExport.ReadXenonExport(ctx, input)
	if err != nil {
		return err
	}
	read.secrets = exported.Secrets
	switch {
	case exported.Workspace != nil:
		read.items = append(read.items, entity.NewWorkspaceImportItem(read.nextKey(), *exported.Workspace))
	case exported.Environment != nil:
		environment := exported.Environment.Duplicated(idutils.NewUlid(), exported.Environment.Name)
		item := entity.NewVariablesImportItem(read.nextKey(), importitemkind.Environment, entity.NewImportVariables(environment, nil))
		read.items = append(read.items, item.FromSource(exported.Environment.ID))
	default:
		read.addSourcedTree(*exported.Root, exported.Root.Reidentified(idutils.NewUlid))
	}
	return nil
}

func translateBrunoScripts(sources []entity.BrunoScripts) (map[string]entity.Scripts, []entity.ImportWarning) {
	scripts := map[string]entity.Scripts{}
	warnings := make([]entity.ImportWarning, 0)
	for _, source := range sources {
		preRequest := usecase.ConvertBrunoPreRequestVars(source.PreRequestVars).Append(usecase.TranslateBrunoScript(source.PreRequest))
		postResponse := usecase.TranslateBrunoScript(source.PostResponse).
			Append(usecase.ConvertBrunoPostResponseVars(source.PostResponseVars)).
			Append(usecase.ConvertBrunoAssertions(source.Assertions)).
			Append(usecase.TranslateBrunoScript(source.Tests))
		translated, untranslated := source.Translated(preRequest, postResponse)
		scripts[source.ItemID] = translated
		warnings = append(warnings, untranslated...)
	}
	return scripts, warnings
}
