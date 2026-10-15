package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/importitemkind"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type importReaders struct {
	readPostmanCollection  usecase.ReadPostmanCollection
	readPostmanEnvironment usecase.ReadPostmanEnvironment
	readPostmanDump        usecase.ReadPostmanDump
	readPostmanV3          usecase.ReadPostmanV3
	readBrunoCollection    usecase.ReadBrunoCollection
	readCurlCommand        usecase.ReadCurlCommand
	readXenonExport        usecase.ReadXenonExport
	readImportFolderFiles  usecase.ReadImportFolderFiles
}

func (i importReaders) detected(ctx context.Context, inputs []entity.ImportInput, tempDirs []string) (*entity.ImportSource, []string, error) {
	detected, expandedDirs, skippedWarnings, err := i.expanded(ctx, inputs)
	if err != nil {
		return nil, expandedDirs, err
	}
	read := entity.NewImportReadout(len(detected))
	for _, input := range detected {
		if err := i.read(ctx, input, &read); err != nil {
			return nil, expandedDirs, err
		}
	}
	read.Warn(skippedWarnings)
	allTempDirs := append(append([]string{}, tempDirs...), expandedDirs...)
	source := entity.NewImportSource(idutils.NewUlid(), detected, allTempDirs, timeutils.Now()).WithItems(read.Items).WithSourceWarnings(read.Warnings).WithEncryptedSecrets(read.EncryptedSecrets)
	return &source, expandedDirs, nil
}

func (i importReaders) expanded(ctx context.Context, inputs []entity.ImportInput) ([]entity.ImportInput, []string, []entity.ImportWarning, error) {
	expanded := make([]entity.ImportInput, 0, len(inputs))
	tempDirs := make([]string, 0)
	warnings := make([]entity.ImportWarning, 0)
	for _, input := range inputs {
		kind, version, err := usecase.DetectImportFormat(input.Probe)
		if err != nil {
			return nil, tempDirs, nil, err
		}
		input = input.Detected(kind, version)
		if !input.IsLooseFolder() {
			expanded = append(expanded, input)
			continue
		}
		listing, dirs, err := i.looseFolder(ctx, input)
		tempDirs = append(tempDirs, dirs...)
		if err != nil {
			return nil, tempDirs, nil, err
		}
		expanded = append(expanded, listing.Inputs...)
		if listing.HasSkipped() {
			warnings = append(warnings, listing.SkippedWarning())
		}
	}
	return expanded, tempDirs, warnings, nil
}

func (i importReaders) looseFolder(ctx context.Context, folder entity.ImportInput) (*entity.ImportFolderListing, []string, error) {
	files, tempDirs, skipped, err := i.readImportFolderFiles.ReadFolderFiles(ctx, folder)
	if err != nil {
		return nil, tempDirs, err
	}
	listing := entity.NewImportFolderListing(folder.Name(), skipped)
	for _, file := range files {
		kind, version, err := usecase.DetectImportFormat(file.Probe)
		file = file.Detected(kind, version)
		if err == nil && file.IsLooseFolder() {
			err = usecase.RejectUnimportableFolder(file.Name())
		}
		if err != nil {
			listing.Skip(file.Name(), err)
			continue
		}
		listing.Add(file)
	}
	if listing.IsEmpty() {
		return nil, tempDirs, usecase.RejectUnimportableFolder(folder.Name())
	}
	return &listing, tempDirs, nil
}

func (i importReaders) read(ctx context.Context, input entity.ImportInput, read *entity.ImportReadout) error {
	switch {
	case input.IsPostmanV3():
		return i.postmanV3(ctx, input, read)
	case input.IsPostmanCollection():
		return i.collection(ctx, input.Document(), read)
	case input.IsPostmanEnvironment():
		return i.variables(ctx, input.Document(), importitemkind.Environment, read)
	case input.IsPostmanGlobals():
		return i.variables(ctx, input.Document(), importitemkind.Globals, read)
	case input.IsPostmanDump():
		return i.dump(ctx, input, read)
	case input.IsBruno():
		return i.bruno(ctx, input, read)
	case input.IsXenon():
		return i.xenon(ctx, input, read)
	default:
		return i.curl(ctx, input, read)
	}
}

func (i importReaders) collection(ctx context.Context, document entity.ImportDocument, read *entity.ImportReadout) error {
	collection, err := i.readPostmanCollection.ReadCollection(ctx, document)
	if err != nil {
		return err
	}
	read.Append(withPmUsages(collection.ImportItem(read.NextKey(), importitemkind.Collection)))
	return nil
}

func (i importReaders) postmanV3(ctx context.Context, input entity.ImportInput, read *entity.ImportReadout) error {
	source, err := i.readPostmanV3.ReadPostmanV3(ctx, input)
	if err != nil {
		return err
	}
	for _, collection := range source.Collections {
		read.Append(withPmUsages(collection.ImportItem(read.NextKey(), importitemkind.Collection)))
	}
	if source.HasRequest() {
		read.Append(withPmUsages(source.Request.ImportItem(read.NextKey(), importitemkind.Request)))
	}
	for _, environment := range source.Environments {
		read.Append(environment.ImportItem(read.NextKey(), importitemkind.Environment))
	}
	if source.HasGlobals() {
		read.Append(source.Globals.ImportItem(read.NextKey(), importitemkind.Globals))
	}
	read.Warn(source.SkippedWarnings())
	return nil
}

func (i importReaders) bruno(ctx context.Context, input entity.ImportInput, read *entity.ImportReadout) error {
	collection, err := i.readBrunoCollection.ReadCollection(ctx, input)
	if err != nil {
		return err
	}
	scripts, warnings := translateBrunoScripts(collection.Scripts)
	read.Append(withPmUsages(entity.NewImportItem(read.NextKey(), importitemkind.Collection, collection.Root.WithScripts(scripts), append(collection.Warnings, warnings...))))
	for _, environment := range collection.Environments {
		read.Append(entity.NewVariablesImportItem(read.NextKey(), importitemkind.Environment, environment))
	}
	return nil
}

func (i importReaders) curl(ctx context.Context, input entity.ImportInput, read *entity.ImportReadout) error {
	parsed, err := i.readCurlCommand.ReadCommand(ctx, input.PastedText())
	if err != nil {
		return err
	}
	request := usecase.DeriveParamsFromUrl(parsed.Request)
	request = request.Identified(idutils.NewUlid(), usecase.DeriveCurlRequestName(request))
	read.Append(withPmUsages(entity.NewImportItem(read.NextKey(), importitemkind.Request, entity.NewRequestImportNode(request), []entity.ImportWarning{})))
	return nil
}

func (i importReaders) variables(ctx context.Context, document entity.ImportDocument, kind enums.ImportItemKind, read *entity.ImportReadout) error {
	variables, err := i.readPostmanEnvironment.ReadEnvironment(ctx, document)
	if err != nil {
		return err
	}
	read.Append(entity.NewVariablesImportItem(read.NextKey(), kind, *variables))
	return nil
}

func (i importReaders) dump(ctx context.Context, input entity.ImportInput, read *entity.ImportReadout) error {
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
	read.Warn(dump.SkippedWarnings())
	return nil
}

func (i importReaders) xenon(ctx context.Context, input entity.ImportInput, read *entity.ImportReadout) error {
	exported, err := i.readXenonExport.ReadXenonExport(ctx, input)
	if err != nil {
		return err
	}
	read.KeepEncryptedSecrets(exported.Secrets)
	switch {
	case exported.IsWorkspace():
		read.Append(entity.NewWorkspaceImportItem(read.NextKey(), *exported.Workspace))
	case exported.IsEnvironment():
		environment := exported.Environment.Duplicated(idutils.NewUlid(), exported.Environment.Name)
		item := entity.NewVariablesImportItem(read.NextKey(), importitemkind.Environment, entity.NewImportVariables(environment, nil))
		read.Append(item.FromSource(exported.Environment.ID))
	default:
		root := exported.Root.Reidentified(idutils.NewUlid)
		item := entity.NewImportItem(read.NextKey(), entity.ImportItemKindOf(root), root, []entity.ImportWarning{})
		read.Append(withPmUsages(item).FromSource(exported.Root.ID()))
	}
	return nil
}

func withPmUsages(item entity.ImportItem) entity.ImportItem {
	usages := usecase.ScanPmUsage(item.Scripts())
	return item.WithPmUsages(usages).WithWarnings(usecase.ListPmUsageWarnings(usages))
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
