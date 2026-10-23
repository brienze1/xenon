package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type startRun struct {
	starter         *runStarter
	planner         *runPlanner
	findAppSettings usecase.FindAppSettings
	approver        *aiApprover
}

func NewStartRunService(
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
	lockLocalVariables usecase.LockLocalVariables,
	loadTree usecase.LoadTree,
	claimActiveRun usecase.ClaimActiveRun,
	changeActiveRun usecase.ChangeActiveRun,
	releaseActiveRun usecase.ReleaseActiveRun,
	findActiveRuns usecase.FindActiveRuns,
	saveRun usecase.SaveRun,
	findRuns usecase.FindRuns,
	publishRunEvent usecase.PublishRunEvent,
	findIterationData usecase.FindIterationData,
	listEnvironments usecase.ListEnvironments,
	deleteRuns usecase.DeleteRuns,
	findRunnerPreferences usecase.FindRunnerPreferences,
	saveRunnerPreferences usecase.SaveRunnerPreferences,
	deleteRunnerPreferences usecase.DeleteRunnerPreferences,
	queueAIApproval usecase.QueueAIApproval,
	settleAIApproval usecase.SettleAIApproval,
	scheduleAIApprovalTimeout usecase.ScheduleAIApprovalTimeout,
	focusWindow usecase.FocusWindow,
	publishAIEvent usecase.PublishAIEvent,
) adapter.StartRunService {
	return &startRun{
		starter: newRunStarter(
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
			loadTree,
			claimActiveRun,
			changeActiveRun,
			releaseActiveRun,
			findActiveRuns,
			saveRun,
			findRuns,
			deleteRuns,
			publishRunEvent,
			findIterationData,
			listEnvironments,
		),
		planner: newRunPlanner(
			findWorkspaceRegistry,
			loadTree,
			findRequest,
			findLocalVariables,
			listEnvironments,
			findActiveRuns,
			deleteRuns,
			releaseResponse,
			findRunnerPreferences,
			saveRunnerPreferences,
			deleteRunnerPreferences,
			findIterationData,
		),
		findAppSettings: findAppSettings,
		approver:        newAIApprover(queueAIApproval, settleAIApproval, scheduleAIApprovalTimeout, focusWindow, publishAIEvent),
	}
}

func (s *startRun) Start(ctx context.Context, start entity.RunStart, appVersion string) (*entity.Run, error) {
	logger.Info(ctx, "Started", start, map[string]any{"appVersion": appVersion})

	run, _, err := s.starter.start(ctx, start, appVersion)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", start, map[string]any{"appVersion": appVersion, "result": run})
	return run, nil
}

func (s *startRun) StartForAI(ctx context.Context, localId, path, itemId string, overrides entity.RunOverrides, appVersion, client, sessionId string) (entity.AIRunStart, error) {
	logger.Info(ctx, "Started", overrides, map[string]any{"localId": localId, "path": path, "itemId": itemId, "appVersion": appVersion, "client": client, "sessionId": sessionId})

	root, tree, err := s.planner.workspaceTree(ctx, localId)
	if err != nil {
		return entity.AIRunStart{}, err
	}
	targetId, target := tree.ResolveRunTarget(path, itemId)
	started := entity.NewAIRunStart(target)
	plan, err := s.planner.plan(ctx, localId, root, *tree, targetId)
	if err != nil {
		return started, err
	}
	config, known := plan.Configured(overrides)
	if !known {
		logger.Info(ctx, "Unknown environment, not started", overrides, map[string]any{"localId": localId, "itemId": targetId})
		return started.WithUnknownEnvironment(), nil
	}
	run, refused, err := s.startForAI(ctx, entity.NewRunStart(localId, targetId, config), appVersion, client, sessionId)
	if err != nil {
		return started, err
	}
	started = started.WithOutcome(run, refused, config.AIChanges())

	logger.Info(ctx, "Finished", overrides, map[string]any{"localId": localId, "path": path, "itemId": itemId, "targetId": targetId, "appVersion": appVersion, "client": client, "sessionId": sessionId, "result": started})
	return started, nil
}

func (s *startRun) startForAI(ctx context.Context, start entity.RunStart, appVersion, client, sessionId string) (*entity.Run, *entity.AIApproval, error) {
	run, confirmation, err := s.starter.start(ctx, start, appVersion)
	if confirmation == nil {
		return run, nil, err
	}
	return s.startConfirmed(ctx, start, appVersion, client, sessionId, *confirmation)
}

func (s *startRun) startConfirmed(ctx context.Context, start entity.RunStart, appVersion, client, sessionId string, confirmation entity.RunConfirmation) (*entity.Run, *entity.AIApproval, error) {
	settings, err := s.findAppSettings.FindAppSettings(ctx)
	if err != nil {
		return nil, nil, err
	}
	if !settings.SkipsAIApprovals() {
		approval, err := s.approver.request(ctx, confirmation.Approval(client, start.Config), sessionId)
		if err != nil {
			return nil, nil, err
		}
		if !approval.IsApproved() {
			return nil, &approval, nil
		}
	}
	run, _, err := s.starter.start(ctx, confirmation.Confirmed(start), appVersion)
	return run, nil, err
}
