package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
)

type sendRequest struct {
	pipeline         *sendPipeline
	saveHistoryEntry usecase.SaveHistoryEntry
	publishAIEvent   usecase.PublishAIEvent
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
	lockLocalVariables usecase.LockLocalVariables,
) adapter.SendRequestService {
	return &sendRequest{
		pipeline: newSendPipeline(
			findWorkspaceRegistry,
			findTreeItem,
			saveTreeItem,
			findRequest,
			loadWorkspace,
			saveWorkspaceFile,
			findEnvironment,
			saveEnvironment,
			findLocalVariables,
			saveLocalVariables,
			findVault,
			saveVault,
			findVaultSettings,
			saveVaultSettings,
			probeKeychain,
			publishVariableEvent,
			findAppSettings,
			runScript,
			appendConsoleEntries,
			clearConsole,
			publishConsoleEntries,
			generateDynamicValue,
			sendHttpRequest,
			saveResponse,
			releaseResponse,
			lockOAuthToken,
			refreshOAuthToken,
			signAwsRequest,
			answerDigest,
			findDigestState,
			saveDigestState,
			manageCookieJar,
			findCookies,
			saveCookies,
			findCookieAllowlist,
			publishWorkspaceEvent,
			registerKnownSecrets,
			lockLocalVariables,
		),
		saveHistoryEntry: saveHistoryEntry,
		publishAIEvent:   publishAIEvent,
	}
}

func (s *sendRequest) Send(ctx context.Context, localId, tabId, aliasId string, request entity.Request, appVersion string, acknowledgement entity.SendAcknowledgement, origin entity.SendOrigin) (*entity.SendExecution, error) {
	logger.Info(ctx, "Started", origin, map[string]any{"localId": localId, "tabId": tabId, "aliasId": aliasId, "request": request, "appVersion": appVersion, "acknowledgement": acknowledgement})

	sent, err := s.pipeline.send(ctx, sendCall{localId: localId, tabId: tabId, aliasId: aliasId, request: request, appVersion: appVersion, acknowledgement: acknowledgement}, entity.NoRunContext())
	if err != nil {
		return nil, err
	}
	ctx = logger.WithRedactions(ctx, sent.secrets...)
	if err := s.saveHistoryEntry.Save(ctx, localId, entity.NewHistoryEntry(idutils.NewUlid(), sent.sentAt, request, *sent.execution, origin).ForAlias(sent.alias), sent.settings.HistoryCaps()); err != nil {
		return nil, err
	}
	if err := s.publishAIEvent.HistoryChanged(ctx, localId); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", origin, map[string]any{"localId": localId, "tabId": tabId, "aliasId": aliasId, "request": request, "appVersion": appVersion, "acknowledgement": acknowledgement, "result": sent.execution})
	return sent.execution, nil
}
