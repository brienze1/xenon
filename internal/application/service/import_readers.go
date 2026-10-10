package service

import (
	"context"
	"strconv"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/importitemkind"
	"github.com/brienze1/xenon/internal/domain/enums/importkind"
)

type importReaders struct {
	readPostmanCollection  usecase.ReadPostmanCollection
	readPostmanEnvironment usecase.ReadPostmanEnvironment
	readPostmanDump        usecase.ReadPostmanDump
}

type importRead struct {
	items    []entity.ImportItem
	warnings []entity.ImportWarning
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
	default:
		return nil
	}
}

func (i importReaders) collection(ctx context.Context, document entity.ImportDocument, read *importRead) error {
	root, warnings, err := i.readPostmanCollection.ReadCollection(ctx, document)
	if err != nil {
		return err
	}
	item := entity.NewImportItem(read.nextKey(), importitemkind.Collection, *root, warnings)
	usages := usecase.ScanPmUsage(item.Scripts())
	read.items = append(read.items, item.WithPmUsages(usages).WithWarnings(usecase.PmUsageWarnings(usages)))
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

func (i *importRead) nextKey() string {
	return strconv.Itoa(len(i.items))
}
