package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/exportitemkind"
	"github.com/brienze1/xenon/internal/domain/enums/exportsecrets"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

const (
	postmanCollectionsFolder  = "collections/"
	postmanEnvironmentsFolder = "environments/"
	postmanGlobalsEntry       = "globals.postman_globals.json"
	postmanCollectionSuffix   = ".postman_collection.json"
	postmanEnvironmentSuffix  = ".postman_environment.json"
)

type postmanExport struct {
	writePostmanExport usecase.WritePostmanExport
	items              exportItems
}

func (p postmanExport) documents(ctx context.Context, root string, request entity.ExportRequest) (*exportDocuments, error) {
	local, err := p.items.local(ctx, request.LocalID)
	if err != nil {
		return nil, err
	}
	documents := &exportDocuments{summary: entity.NewExportSummary(request.TargetPath, exportsecrets.NamesOnly)}
	switch request.ItemKind {
	case exportitemkind.Workspace:
		err = p.workspace(ctx, root, request, *local, documents)
	case exportitemkind.Environment:
		err = p.environmentById(ctx, root, request, *local, documents)
	case exportitemkind.Globals:
		err = p.globals(ctx, root, request, *local, documents)
	default:
		err = p.item(ctx, root, request, *local, documents)
	}
	if err != nil {
		return nil, err
	}
	return documents, nil
}

func (p postmanExport) item(ctx context.Context, root string, request entity.ExportRequest, local entity.LocalVariables, documents *exportDocuments) error {
	node, ancestors, err := p.items.item(ctx, root, request.ItemID)
	if err != nil {
		return err
	}
	return p.collection(ctx, node, ancestors, request, local, documents)
}

func (p postmanExport) collection(ctx context.Context, node entity.ExportNode, ancestors []entity.Collection, request entity.ExportRequest, local entity.LocalVariables, documents *exportDocuments) error {
	exported := usecase.EffectiveExportRoot(node, ancestors)
	if request.CurrentValuesAsInitial {
		scope := exported.Container.ScopeRef()
		if collection, found := entity.RootCollection(ancestors); found {
			scope = collection.ScopeRef()
		}
		exported.Container = exported.Container.WithVariables(usecase.ApplyCurrentValuesAsInitial(exported.Container.Variables, local.CurrentValues(scope)))
	}
	data, err := p.writePostmanExport.WriteCollection(ctx, exported)
	if err != nil {
		return err
	}
	documents.add(postmanCollectionsFolder, node.Name(), postmanCollectionSuffix, data)
	drops := append(usecase.FindPostmanDrops(node), usecase.FindPostmanSecretVariableDrops(exported.Container.Variables, exported.Path)...)
	documents.summary = documents.summary.With(exported.Counts(), drops, usecase.FindLiteralCredentials(node))
	return nil
}

func (p postmanExport) environmentById(ctx context.Context, root string, request entity.ExportRequest, local entity.LocalVariables, documents *exportDocuments) error {
	environment, err := p.items.findEnvironment.Find(ctx, root, request.ItemID)
	if err != nil {
		return err
	}
	return p.environment(ctx, *environment, request, local, documents)
}

func (p postmanExport) environment(ctx context.Context, environment entity.Environment, request entity.ExportRequest, local entity.LocalVariables, documents *exportDocuments) error {
	if request.CurrentValuesAsInitial {
		environment = environment.WithVariables(usecase.ApplyCurrentValuesAsInitial(environment.Variables, local.CurrentValues(environment.ScopeRef())))
	}
	data, err := p.writePostmanExport.WriteEnvironment(ctx, environment)
	if err != nil {
		return err
	}
	documents.add(postmanEnvironmentsFolder, environment.Name, postmanEnvironmentSuffix, data)
	counts := entity.ExportCounts{Environments: 1, Variables: len(environment.Variables)}
	documents.summary = documents.summary.With(counts, usecase.FindPostmanEnvironmentDrops(environment, environment.Name), nil)
	return nil
}

func (p postmanExport) globals(ctx context.Context, root string, request entity.ExportRequest, local entity.LocalVariables, documents *exportDocuments) error {
	workspace, err := p.items.loadWorkspace.Load(ctx, root)
	if err != nil {
		return err
	}
	if request.CurrentValuesAsInitial {
		*workspace = workspace.WithVariables(usecase.ApplyCurrentValuesAsInitial(workspace.Variables, local.CurrentValues(entity.NewGlobalsScopeRef())))
	}
	data, err := p.writePostmanExport.WriteGlobals(ctx, *workspace)
	if err != nil {
		return err
	}
	documents.entries = append(documents.entries, entity.NewExportEntry(postmanGlobalsEntry, data))
	documents.summary = documents.summary.With(entity.ExportCounts{Variables: len(workspace.Variables)}, nil, nil)
	return nil
}

func (p postmanExport) workspace(ctx context.Context, root string, request entity.ExportRequest, local entity.LocalVariables, documents *exportDocuments) error {
	collections, err := p.items.collections(ctx, root)
	if err != nil {
		return err
	}
	for _, collection := range collections {
		if err := p.collection(ctx, collection, nil, request, local, documents); err != nil {
			return err
		}
	}
	environments, err := p.items.listEnvironments.List(ctx, root)
	if err != nil {
		return err
	}
	for _, environment := range environments {
		if environment.IsMalformed() {
			continue
		}
		if err := p.environment(ctx, environment, request, local, documents); err != nil {
			return err
		}
	}
	return p.globals(ctx, root, request, local, documents)
}

func (d *exportDocuments) add(folder, name, suffix string, data []byte) {
	base := usecase.DeriveFileName(name, itemkind.Folder, d.names)
	d.names = append(d.names, base)
	d.entries = append(d.entries, entity.NewExportEntry(folder+base+suffix, data))
}
