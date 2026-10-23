package service

import (
	"context"
	"slices"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageAlias struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findRequest           usecase.FindRequest
	saveRequestAliases    usecase.SaveRequestAliases
	listEnvironments      usecase.ListEnvironments
	tokens                oauthTokens
}

func NewManageAliasService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findRequest usecase.FindRequest,
	saveRequestAliases usecase.SaveRequestAliases,
	findVault usecase.FindVault,
	saveVault usecase.SaveVault,
	findVaultSettings usecase.FindVaultSettings,
	saveVaultSettings usecase.SaveVaultSettings,
	probeKeychain usecase.ProbeKeychain,
	publishVariableEvent usecase.PublishVariableEvent,
	listEnvironments usecase.ListEnvironments,
) adapter.ManageAliasService {
	return &manageAlias{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findRequest:           findRequest,
		saveRequestAliases:    saveRequestAliases,
		listEnvironments:      listEnvironments,
		tokens: oauthTokens{
			vaults: vaultWriter{
				findVault:            findVault,
				saveVault:            saveVault,
				findVaultSettings:    findVaultSettings,
				saveVaultSettings:    saveVaultSettings,
				probeKeychain:        probeKeychain,
				publishVariableEvent: publishVariableEvent,
			},
		},
	}
}

func (m *manageAlias) Create(ctx context.Context, localId, requestId, name string, environmentIds []string, baseRevision string, force bool) (*entity.Request, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "requestId": requestId, "name": name, "environmentIds": environmentIds, "baseRevision": baseRevision, "force": force})

	loaded, err := m.load(ctx, localId, requestId, baseRevision, force)
	if err != nil {
		return nil, err
	}
	created := entity.NewRequestAlias(idutils.NewUlid(), name, loaded.request, environmentIds)
	if err := m.checkEnvironments(ctx, loaded.root, created.EnvironmentIDs); err != nil {
		return nil, err
	}
	saved, err := m.save(ctx, *loaded, created, loaded.request.WithAliasAppended(created))
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "requestId": requestId, "name": name, "environmentIds": environmentIds, "baseRevision": baseRevision, "force": force, "result": *saved})
	return saved, nil
}

func (m *manageAlias) Save(ctx context.Context, localId, requestId string, alias entity.RequestAlias, baseRevision string, force bool) (*entity.Request, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "requestId": requestId, "alias": alias, "baseRevision": baseRevision, "force": force})

	loaded, err := m.load(ctx, localId, requestId, baseRevision, force)
	if err != nil {
		return nil, err
	}
	stored, err := usecase.ResolveAlias(loaded.request, alias.ID)
	if err != nil {
		return nil, err
	}
	edited := stored.Edited(alias.Version.WithFileRefsRelativeTo(loaded.root)).Renamed(alias.Name)
	saved, err := m.save(ctx, *loaded, edited, loaded.request.WithAlias(edited))
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "requestId": requestId, "alias": alias, "baseRevision": baseRevision, "force": force, "result": *saved})
	return saved, nil
}

func (m *manageAlias) SetEnvironments(ctx context.Context, localId, requestId, aliasId string, environmentIds []string, baseRevision string, force bool) (*entity.Request, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "requestId": requestId, "aliasId": aliasId, "environmentIds": environmentIds, "baseRevision": baseRevision, "force": force})

	loaded, err := m.load(ctx, localId, requestId, baseRevision, force)
	if err != nil {
		return nil, err
	}
	stored, err := usecase.ResolveAlias(loaded.request, aliasId)
	if err != nil {
		return nil, err
	}
	linked := stored.WithEnvironments(environmentIds)
	if err := m.checkEnvironments(ctx, loaded.root, linked.EnvironmentIDs); err != nil {
		return nil, err
	}
	saved, err := m.saveRequestAliases.SaveAliases(ctx, loaded.root, loaded.request.WithAlias(linked))
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "requestId": requestId, "aliasId": aliasId, "environmentIds": environmentIds, "baseRevision": baseRevision, "force": force, "result": *saved})
	return saved, nil
}

func (m *manageAlias) Delete(ctx context.Context, localId, requestId, aliasId, baseRevision string, force bool) (*entity.Request, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "requestId": requestId, "aliasId": aliasId, "baseRevision": baseRevision, "force": force})

	loaded, err := m.load(ctx, localId, requestId, baseRevision, force)
	if err != nil {
		return nil, err
	}
	if _, err := usecase.ResolveAlias(loaded.request, aliasId); err != nil {
		return nil, err
	}
	saved, err := m.saveRequestAliases.SaveAliases(ctx, loaded.root, loaded.request.WithoutAlias(aliasId))
	if err != nil {
		return nil, err
	}
	if err := m.tokens.delete(ctx, localId, entity.AliasAuthOwner(requestId, aliasId)); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "requestId": requestId, "aliasId": aliasId, "baseRevision": baseRevision, "force": force, "result": *saved})
	return saved, nil
}

func (m *manageAlias) load(ctx context.Context, localId, requestId, baseRevision string, force bool) (*loadedRequest, error) {
	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	current, err := m.findRequest.FindRevision(ctx, workspace.Path, requestId)
	if err != nil {
		return nil, err
	}
	if current.BlocksSave(baseRevision, force) {
		return nil, usecase.RejectStaleAliasWrite()
	}
	request, err := m.findRequest.Find(ctx, workspace.Path, requestId)
	if err != nil {
		return nil, err
	}
	return &loadedRequest{root: workspace.Path, request: *request}, nil
}

func (m *manageAlias) checkEnvironments(ctx context.Context, root string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	environments, err := m.listEnvironments.List(ctx, root)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if !slices.ContainsFunc(environments, func(environment entity.Environment) bool { return environment.ID == id && environment.Error == "" }) {
			return usecase.RejectUnknownEnvironment()
		}
	}
	return nil
}

func (m *manageAlias) save(ctx context.Context, loaded loadedRequest, alias entity.RequestAlias, updated entity.Request) (*entity.Request, error) {
	switch {
	case !alias.HasValidName():
		return nil, usecase.RejectInvalidAliasName()
	case alias.HasReservedName():
		return nil, usecase.RejectReservedAliasName()
	case updated.HasAliasNamed(alias.Name, alias.ID):
		return nil, usecase.RejectTakenAliasName()
	}
	return m.saveRequestAliases.SaveAliases(ctx, loaded.root, updated)
}
