package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/exportitemkind"
)

type xenonExport struct {
	writeXenonExport     usecase.WriteXenonExport
	encryptExportSecrets usecase.EncryptExportSecrets
	findVault            usecase.FindVault
	findVaultSettings    usecase.FindVaultSettings
	items                exportItems
}

func (x xenonExport) documents(ctx context.Context, root string, request entity.ExportRequest) (*exportDocuments, error) {
	export, counts, err := x.exported(ctx, root, request)
	if err != nil {
		return nil, err
	}
	_, scopes, err := x.items.scopes(ctx, root, request.ItemKind, request.ItemID)
	if err != nil {
		return nil, err
	}
	scopes = request.ExportedScopes(scopes)
	if request.CurrentValuesAsInitial {
		local, err := x.items.local(ctx, request.LocalID)
		if err != nil {
			return nil, err
		}
		export = export.WithCurrentValues(scopes, *local)
	}
	if request.IncludeSecrets {
		if export, err = x.withSecrets(ctx, export, request, scopes); err != nil {
			return nil, err
		}
	}
	entries, err := x.writeXenonExport.WriteXenonExport(ctx, root, export)
	if err != nil {
		return nil, err
	}
	return &exportDocuments{entries: entries, summary: entity.NewExportSummary(request.TargetPath, export.ExportedSecrets()).With(counts, nil, nil)}, nil
}

func (x xenonExport) exported(ctx context.Context, root string, request entity.ExportRequest) (entity.XenonExport, entity.ExportCounts, error) {
	export := entity.NewXenonExport(request)
	switch request.ItemKind {
	case exportitemkind.Workspace:
		counts, err := x.workspaceCounts(ctx, root)
		return export, counts, err
	case exportitemkind.Environment:
		environment, err := x.items.findEnvironment.Find(ctx, root, request.ItemID)
		if err != nil {
			return export, entity.ExportCounts{}, err
		}
		return export, environment.ExportCounts(), nil
	default:
		node, _, err := x.items.item(ctx, root, request.ItemID)
		if err != nil {
			return export, entity.ExportCounts{}, err
		}
		return export.WithNode(node), node.Counts(), nil
	}
}

func (x xenonExport) workspaceCounts(ctx context.Context, root string) (entity.ExportCounts, error) {
	counts := entity.ExportCounts{}
	collections, err := x.items.collections(ctx, root)
	if err != nil {
		return counts, err
	}
	for _, collection := range collections {
		counts = counts.Add(collection.Counts())
	}
	environments, err := x.items.listEnvironments.List(ctx, root)
	if err != nil {
		return counts, err
	}
	for _, environment := range entity.WellFormedEnvironments(environments) {
		counts = counts.Add(environment.ExportCounts())
	}
	return counts, nil
}

func (x xenonExport) withSecrets(ctx context.Context, export entity.XenonExport, request entity.ExportRequest, scopes []entity.ExportScope) (entity.XenonExport, error) {
	if !entity.ExportScopesHaveSecrets(scopes) {
		return export, nil
	}
	vault, err := x.findVault.Find(ctx, request.LocalID)
	if err != nil {
		return export, err
	}
	settings, err := x.findVaultSettings.FindSettings(ctx)
	if err != nil {
		return export, err
	}
	if err := usecase.CheckVaultReadable(*vault, *settings); err != nil {
		return export, err
	}
	bundle := entity.NewExportSecretBundle(scopes, *vault)
	if bundle.IsEmpty() {
		return export, nil
	}
	armored, err := x.encryptExportSecrets.Encrypt(ctx, bundle, request.Passphrase)
	if err != nil {
		return export, err
	}
	return export.WithSecrets(armored), nil
}
