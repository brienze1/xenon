package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type sendRequest struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findRequest           usecase.FindRequest
	findAppSettings       usecase.FindAppSettings
	runScript             usecase.RunScript
	sendHttpRequest       usecase.SendHttpRequest
	saveResponse          usecase.SaveResponse
	saveHistoryEntry      usecase.SaveHistoryEntry
	registerKnownSecrets  usecase.RegisterKnownSecrets
	publishAIEvent        usecase.PublishAIEvent
	generateDynamicValue  usecase.GenerateDynamicValue
	signAwsRequest        usecase.SignAwsRequest
	answerDigest          usecase.AnswerDigestChallenge
	findDigestState       usecase.FindDigestState
	saveDigestState       usecase.SaveDigestState
	findCookieAllowlist   usecase.FindCookieAllowlist
	listEnvironments      usecase.ListEnvironments
	tokens                oauthTokens
	cookies               cookieJar
	variables             requestVariables
	scriptVariables       scriptVariables
	console               consoleLog
}

func NewSendRequestService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	saveTreeItem usecase.SaveTreeItem,
	findRequest usecase.FindRequest,
	loadWorkspace usecase.LoadWorkspace,
	saveWorkspaceFile usecase.SaveWorkspaceFile,
	findEnvironment usecase.FindEnvironment,
	saveEnvironment usecase.SaveEnvironment,
	findLocalVariables usecase.FindLocalVariables,
	saveLocalVariables usecase.SaveLocalVariables,
	findVault usecase.FindVault,
	saveVault usecase.SaveVault,
	findVaultSettings usecase.FindVaultSettings,
	saveVaultSettings usecase.SaveVaultSettings,
	probeKeychain usecase.ProbeKeychain,
	publishVariableEvent usecase.PublishVariableEvent,
	findAppSettings usecase.FindAppSettings,
	runScript usecase.RunScript,
	appendConsoleEntries usecase.AppendConsoleEntries,
	clearConsole usecase.ClearConsole,
	publishConsoleEntries usecase.PublishConsoleEntries,
	generateDynamicValue usecase.GenerateDynamicValue,
	sendHttpRequest usecase.SendHttpRequest,
	saveResponse usecase.SaveResponse,
	releaseResponse usecase.ReleaseResponse,
	saveHistoryEntry usecase.SaveHistoryEntry,
	lockOAuthToken usecase.LockOAuthToken,
	refreshOAuthToken usecase.RefreshOAuthToken,
	signAwsRequest usecase.SignAwsRequest,
	answerDigest usecase.AnswerDigestChallenge,
	findDigestState usecase.FindDigestState,
	saveDigestState usecase.SaveDigestState,
	manageCookieJar usecase.ManageCookieJar,
	findCookies usecase.FindCookies,
	saveCookies usecase.SaveCookies,
	findCookieAllowlist usecase.FindCookieAllowlist,
	publishWorkspaceEvent usecase.PublishWorkspaceEvent,
	registerKnownSecrets usecase.RegisterKnownSecrets,
	publishAIEvent usecase.PublishAIEvent,
	listEnvironments usecase.ListEnvironments,
) adapter.SendRequestService {
	scopes := variableScopes{findEnvironment: findEnvironment, findTreeItem: findTreeItem, loadWorkspace: loadWorkspace}
	vaults := vaultWriter{
		findVault:            findVault,
		saveVault:            saveVault,
		findVaultSettings:    findVaultSettings,
		saveVaultSettings:    saveVaultSettings,
		probeKeychain:        probeKeychain,
		publishVariableEvent: publishVariableEvent,
	}
	return &sendRequest{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findRequest:           findRequest,
		findAppSettings:       findAppSettings,
		runScript:             runScript,
		sendHttpRequest:       sendHttpRequest,
		saveResponse:          saveResponse,
		saveHistoryEntry:      saveHistoryEntry,
		registerKnownSecrets:  registerKnownSecrets,
		publishAIEvent:        publishAIEvent,
		generateDynamicValue:  generateDynamicValue,
		signAwsRequest:        signAwsRequest,
		answerDigest:          answerDigest,
		findDigestState:       findDigestState,
		saveDigestState:       saveDigestState,
		findCookieAllowlist:   findCookieAllowlist,
		listEnvironments:      listEnvironments,
		tokens:                oauthTokens{vaults: vaults, lockOAuthToken: lockOAuthToken, refreshOAuthToken: refreshOAuthToken},
		cookies: cookieJar{
			manageCookieJar:       manageCookieJar,
			findCookies:           findCookies,
			saveCookies:           saveCookies,
			publishWorkspaceEvent: publishWorkspaceEvent,
			vaults:                vaults,
		},
		variables: requestVariables{
			findTreeItem:       findTreeItem,
			loadWorkspace:      loadWorkspace,
			findEnvironment:    findEnvironment,
			findLocalVariables: findLocalVariables,
			findVault:          findVault,
		},
		scriptVariables: scriptVariables{
			scopes:               scopes,
			rowStore:             variableRowStore{scopes: scopes, saveEnvironment: saveEnvironment, saveTreeItem: saveTreeItem, saveWorkspaceFile: saveWorkspaceFile},
			findLocalVariables:   findLocalVariables,
			saveLocalVariables:   saveLocalVariables,
			vaults:               vaults,
			publishVariableEvent: publishVariableEvent,
		},
		console: consoleLog{
			appendConsoleEntries:  appendConsoleEntries,
			clearConsole:          clearConsole,
			publishConsoleEntries: publishConsoleEntries,
			saveResponse:          saveResponse,
			releaseResponse:       releaseResponse,
		},
	}
}

func (s *sendRequest) Send(ctx context.Context, localId, tabId, aliasId string, request entity.Request, appVersion string, acknowledgement entity.SendAcknowledgement, origin entity.SendOrigin) (*entity.SendExecution, error) {
	logger.Info(ctx, "Started", origin, map[string]any{"localId": localId, "tabId": tabId, "aliasId": aliasId, "request": request, "appVersion": appVersion, "acknowledgement": acknowledgement})

	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	scopes, err := s.variables.load(ctx, localId, workspace.Path, request.ID)
	if err != nil {
		return nil, err
	}
	alias, err := s.alias(ctx, workspace.Path, request.ID, aliasId)
	if err != nil {
		return nil, err
	}
	if !alias.IsAvailableIn(scopes.environment) {
		return nil, s.unavailableAlias(ctx, workspace.Path, alias)
	}
	effective := usecase.ResolveEffectiveRequest(request, scopes.ancestors, workspace.Path).ForAlias(request.ID, aliasId).WithAuthOwner(tabId)
	preflight := usecase.CheckSendPreflight(request.Inheriting(effective), scopes.context, scopes.environment)
	if preflight.AwaitsConfirmation(acknowledgement) {
		return nil, usecase.RejectUnconfirmedSend(preflight)
	}
	if preflight.AwaitsSecretsAcknowledgement(acknowledgement) {
		return nil, usecase.RejectUnavailableSecrets(preflight)
	}
	token, err := s.oauthToken(ctx, localId, effective, scopes.context, acknowledgement)
	if err != nil {
		return nil, err
	}
	owners, err := s.scriptOwners(ctx, workspace.Path, request, aliasId, scopes.ancestors)
	if err != nil {
		return nil, err
	}
	settings, err := s.findAppSettings.FindAppSettings(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.cookies.load(ctx, localId); err != nil {
		return nil, err
	}
	cookiesBefore := s.cookies.revision(ctx, localId)
	secrets := append(scopes.context.SecretValues(), token.SecretsOrNone()...)
	ctx = logger.WithRedactions(ctx, secrets...)
	sentAt := timeutils.Now()
	execution := entity.NewSendExecution(idutils.NewUlid())
	execution.Start(request, scopes.scripts, scopes.vault)
	run := sendRun{service: s, scopes: scopes, token: token, localId: localId, tabId: tabId, aliasId: aliasId, root: workspace.Path, appVersion: appVersion, sendId: execution.ID, secrets: secrets, settings: *settings}
	if err := run.execute(ctx, tabId, request, scopes.ancestors, owners, &execution); err != nil {
		return nil, err
	}
	if err := s.cookies.settle(ctx, localId, cookiesBefore); err != nil {
		return nil, err
	}
	s.registerKnownSecrets.Register(ctx, append(secrets, execution.DerivedSecrets(secrets)...)...)
	if err := s.saveHistoryEntry.Save(ctx, localId, entity.NewHistoryEntry(idutils.NewUlid(), sentAt, request, execution, origin).ForAlias(alias).InEnvironment(scopes.environment), settings.HistoryCaps()); err != nil {
		return nil, err
	}
	if err := s.publishAIEvent.HistoryChanged(ctx, localId); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", origin, map[string]any{"localId": localId, "tabId": tabId, "aliasId": aliasId, "request": request, "appVersion": appVersion, "acknowledgement": acknowledgement, "result": &execution})
	return &execution, nil
}

func (s *sendRequest) oauthToken(ctx context.Context, localId string, effective entity.EffectiveRequest, variables entity.VariableContext, acknowledgement entity.SendAcknowledgement) (*entity.OAuth2Token, error) {
	if !effective.Auth.IsOAuth2() {
		return nil, nil
	}
	config := effective.Auth.OAuth2.MapText(func(text string) string {
		return variables.Resolve(text, nil)
	})
	token, expired, err := s.tokens.current(ctx, localId, effective.AuthOwner, config)
	if err != nil {
		return nil, err
	}
	if acknowledgement.RejectsExpiredToken(expired) {
		return nil, usecase.RejectExpiredToken(effective.AuthOwner)
	}
	return token, nil
}

func (s *sendRequest) scriptOwners(ctx context.Context, root string, request entity.Request, aliasId string, ancestors []entity.Collection) ([]entity.ScriptOwner, error) {
	if request.UsesOwnScripts() {
		return request.ScriptOwners(ancestors), nil
	}
	stored, err := s.findRequest.Find(ctx, root, request.ID)
	if err != nil {
		return nil, err
	}
	return request.WithScriptsOf(stored.Version(aliasId)).ScriptOwners(ancestors), nil
}

func (s *sendRequest) unavailableAlias(ctx context.Context, root string, alias entity.RequestAlias) error {
	environments, err := s.listEnvironments.List(ctx, root)
	if err != nil {
		return err
	}
	return usecase.RejectUnavailableAlias(alias.Name, entity.AliasEnvironmentLabels(alias.EnvironmentIDs, environments))
}

func (s *sendRequest) alias(ctx context.Context, root, requestId, aliasId string) (entity.RequestAlias, error) {
	if entity.IsDefaultAliasID(aliasId) {
		return entity.RequestAlias{}, nil
	}
	stored, err := s.findRequest.Find(ctx, root, requestId)
	if err != nil {
		return entity.RequestAlias{}, err
	}
	alias, err := usecase.ResolveAlias(*stored, aliasId)
	if err != nil {
		return entity.RequestAlias{}, err
	}
	return *alias, nil
}
