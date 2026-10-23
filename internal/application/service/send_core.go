package service

import (
	"context"
	"time"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type sendCall struct {
	localId         string
	tabId           string
	aliasId         string
	appVersion      string
	acknowledgement entity.SendAcknowledgement
	request         entity.Request
}

type sentRequest struct {
	sentAt    time.Time
	execution *entity.SendExecution
	secrets   []string
	alias     entity.RequestAlias
	settings  entity.AppSettings
}

func (s *sendRequest) send(ctx context.Context, call sendCall, runContext entity.RunContext) (*sentRequest, error) {
	localId, tabId, aliasId, request := call.localId, call.tabId, call.aliasId, call.request
	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	scopes, err := s.variables.loadIn(ctx, localId, workspace.Path, request.ID, runContext.EnvironmentID)
	if err != nil {
		return nil, err
	}
	scopes = scopes.inRun(runContext)
	alias, err := s.alias(ctx, workspace.Path, request.ID, aliasId)
	if err != nil {
		return nil, err
	}
	effective := usecase.ResolveEffectiveRequest(request, scopes.ancestors, workspace.Path).ForAlias(request.ID, aliasId).WithAuthOwner(tabId)
	preflight := usecase.CheckSendPreflight(request.Inheriting(effective), scopes.context, scopes.environment)
	if preflight.AwaitsConfirmation(call.acknowledgement) {
		return nil, usecase.RejectUnconfirmedSend(preflight)
	}
	if preflight.AwaitsSecretsAcknowledgement(call.acknowledgement) {
		return nil, usecase.RejectUnavailableSecrets(preflight)
	}
	token, err := s.oauthToken(ctx, localId, effective, scopes.context, call.acknowledgement)
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
	jarId := runContext.JarID(localId)
	if err := s.loadWorkspaceJar(ctx, localId, runContext); err != nil {
		return nil, err
	}
	cookiesBefore := s.cookies.revision(ctx, jarId)
	secrets := append(scopes.context.SecretValues(), token.SecretsOrNone()...)
	ctx = logger.WithRedactions(ctx, secrets...)
	sentAt := timeutils.Now()
	execution := entity.NewSendExecution(idutils.NewUlid())
	execution.Start(request, scopes.scripts, scopes.vault)
	run := sendRun{service: s, scopes: scopes, token: token, runContext: runContext, localId: localId, jarId: jarId, tabId: tabId, aliasId: aliasId, root: workspace.Path, appVersion: call.appVersion, sendId: execution.ID, secrets: secrets, settings: *settings}
	if err := run.execute(ctx, tabId, request, scopes.ancestors, owners, &execution); err != nil {
		return nil, err
	}
	if runContext.UsesWorkspaceJar(localId) {
		if err := s.cookies.settle(ctx, localId, cookiesBefore); err != nil {
			return nil, err
		}
	}
	known := append(secrets, execution.DerivedSecrets(secrets)...)
	s.registerKnownSecrets.Register(ctx, known...)
	return &sentRequest{sentAt: sentAt, execution: &execution, alias: alias, settings: *settings, secrets: append(known, run.secretsOf(&execution)...)}, nil
}

func (s *sendRequest) loadWorkspaceJar(ctx context.Context, localId string, runContext entity.RunContext) error {
	if !runContext.UsesWorkspaceJar(localId) {
		return nil
	}
	return s.cookies.load(ctx, localId)
}
