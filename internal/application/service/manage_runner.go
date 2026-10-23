package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageRunner struct {
	core                  *sendRequest
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findRequest           usecase.FindRequest
	findLocalVariables    usecase.FindLocalVariables
	listEnvironments      usecase.ListEnvironments
	loadTree              usecase.LoadTree
	manageCookieJar       usecase.ManageCookieJar
	manageActiveRuns      usecase.ManageActiveRuns
	saveRun               usecase.SaveRun
	findRuns              usecase.FindRuns
	deleteRuns            usecase.DeleteRuns
	publishRunEvent       usecase.PublishRunEvent
	findRunnerPreferences usecase.FindRunnerPreferences
	saveRunnerPreferences usecase.SaveRunnerPreferences
	findIterationData     usecase.FindIterationData
	findResponse          usecase.FindResponse
	releaseResponse       usecase.ReleaseResponse
	saveRunExport         usecase.SaveRunExport
}

func NewManageRunnerService(
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
	loadTree usecase.LoadTree,
	manageActiveRuns usecase.ManageActiveRuns,
	saveRun usecase.SaveRun,
	findRuns usecase.FindRuns,
	publishRunEvent usecase.PublishRunEvent,
	deleteRuns usecase.DeleteRuns,
	listEnvironments usecase.ListEnvironments,
	findRunnerPreferences usecase.FindRunnerPreferences,
	saveRunnerPreferences usecase.SaveRunnerPreferences,
	findIterationData usecase.FindIterationData,
	findResponse usecase.FindResponse,
	saveRunExport usecase.SaveRunExport,
) adapter.ManageRunnerService {
	return &manageRunner{
		core: newSendRequest(
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
			saveHistoryEntry,
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
			publishAIEvent,
			lockLocalVariables,
		),
		findWorkspaceRegistry: findWorkspaceRegistry,
		findRequest:           findRequest,
		findLocalVariables:    findLocalVariables,
		loadTree:              loadTree,
		manageCookieJar:       manageCookieJar,
		manageActiveRuns:      manageActiveRuns,
		saveRun:               saveRun,
		findRuns:              findRuns,
		publishRunEvent:       publishRunEvent,
		deleteRuns:            deleteRuns,
		listEnvironments:      listEnvironments,
		findRunnerPreferences: findRunnerPreferences,
		saveRunnerPreferences: saveRunnerPreferences,
		findIterationData:     findIterationData,
		findResponse:          findResponse,
		releaseResponse:       releaseResponse,
		saveRunExport:         saveRunExport,
	}
}

func (m *manageRunner) Start(ctx context.Context, start entity.RunStart, appVersion string) (*entity.Run, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": start.LocalID, "itemId": start.ItemID, "config": start.Config})

	start.Config = start.Config.WithProfileDefaults()
	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, start.LocalID)
	if err != nil {
		return nil, err
	}
	target, rows, err := m.target(ctx, start, workspace.Path)
	if err != nil {
		return nil, err
	}
	if err := start.Config.Validate(); err != nil {
		return nil, err
	}
	data, err := m.iterationData(ctx, start.Config.DataFile())
	if err != nil {
		return nil, err
	}
	start.Config = start.Config.WithDataIterations(data.Count())
	environment, err := m.environment(ctx, start, workspace.Path)
	if err != nil {
		return nil, err
	}
	if err := m.preflight(ctx, start, workspace.Path, environment.id, rows); err != nil {
		return nil, err
	}
	run, err := m.launch(ctx, start, functionalRun{target: target, rows: rows, root: workspace.Path, environment: environment, data: *data, appVersion: appVersion})
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": start.LocalID, "itemId": start.ItemID, "runId": run.ID})
	return run, nil
}

func (m *manageRunner) Get(ctx context.Context, localId, runId string) (*entity.Run, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	run, err := m.run(ctx, localId, runId)
	if err != nil {
		return nil, err
	}
	released, err := m.releasedResponses(ctx, *run)
	if err != nil {
		return nil, err
	}
	held := run.WithoutResponses(released)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": run.Status})
	return &held, nil
}

func (m *manageRunner) Active(ctx context.Context, localId string) ([]entity.Run, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	active := m.manageActiveRuns.ListWorkspace(ctx, localId)
	runs := make([]entity.Run, 0, len(active))
	for _, run := range active {
		runs = append(runs, run.Run)
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runs": len(runs)})
	return runs, nil
}

func (m *manageRunner) Pause(ctx context.Context, localId, runId string) (*entity.Run, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	run, err := m.steer(ctx, localId, runId, func(run *entity.Run) bool { return run.Pause() }, (*entity.RunControl).Pause)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": run.Status})
	return run, nil
}

func (m *manageRunner) Resume(ctx context.Context, localId, runId string) (*entity.Run, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	run, err := m.steer(ctx, localId, runId, func(run *entity.Run) bool { return run.Resume() }, (*entity.RunControl).Resume)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": run.Status})
	return run, nil
}

func (m *manageRunner) Stop(ctx context.Context, localId, runId string) (*entity.Run, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	active, found := m.manageActiveRuns.Find(ctx, localId, runId)
	if !found {
		return m.Get(ctx, localId, runId)
	}
	active.Control.Stop()

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId})
	return &active.Run, nil
}

func (m *manageRunner) Abandon(ctx context.Context) {
	logger.Info(ctx, "Started")

	active := m.manageActiveRuns.ListActive(ctx)
	for _, run := range active {
		run.Control.Abandon()
	}
	awaitRuns(ctx, active)

	logger.Info(ctx, "Finished", map[string]any{"runs": len(active)})
}

func (m *manageRunner) steer(ctx context.Context, localId, runId string, change func(*entity.Run) bool, steer func(*entity.RunControl)) (*entity.Run, error) {
	active, found := m.manageActiveRuns.Find(ctx, localId, runId)
	if !found {
		return nil, usecase.RejectUnknownRun(runId)
	}
	if active.Run.Config.IsPerformance() {
		return nil, usecase.RejectInvalidRunSetting("type", "a performance run cannot pause")
	}
	changed := false
	steered, found := m.manageActiveRuns.Change(ctx, localId, runId, func(run *entity.Run) {
		if changed = change(run); changed {
			steer(active.Control)
		}
	})
	if !found {
		return m.Get(ctx, localId, runId)
	}
	if changed {
		m.publishRunEvent.Progress(ctx, steered.Run)
	}
	return &steered.Run, nil
}
