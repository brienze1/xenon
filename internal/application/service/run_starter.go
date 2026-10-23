package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

type runEnvironment struct {
	id   string
	name string
}

type runStarter struct {
	pipeline              *sendPipeline
	executor              *runExecutor
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findRequest           usecase.FindRequest
	findLocalVariables    usecase.FindLocalVariables
	loadTree              usecase.LoadTree
	findIterationData     usecase.FindIterationData
}

func newRunStarter(
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
	deleteRuns usecase.DeleteRuns,
	publishRunEvent usecase.PublishRunEvent,
	findIterationData usecase.FindIterationData,
	listEnvironments usecase.ListEnvironments,
) *runStarter {
	pipeline := newSendPipeline(
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
		listEnvironments,
	)
	return &runStarter{
		pipeline:              pipeline,
		executor:              newRunExecutor(pipeline, findRequest, manageCookieJar, claimActiveRun, changeActiveRun, releaseActiveRun, findActiveRuns, saveRun, findRuns, deleteRuns, publishRunEvent, releaseResponse),
		findWorkspaceRegistry: findWorkspaceRegistry,
		findRequest:           findRequest,
		findLocalVariables:    findLocalVariables,
		loadTree:              loadTree,
		findIterationData:     findIterationData,
	}
}

func (s *runStarter) start(ctx context.Context, start entity.RunStart, appVersion string) (*entity.Run, *entity.RunConfirmation, error) {
	start.Config = start.Config.WithProfileDefaults()
	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, start.LocalID)
	if err != nil {
		return nil, nil, err
	}
	target, rows, err := s.target(ctx, start, workspace.Path)
	if err != nil {
		return nil, nil, err
	}
	if err := start.Config.Validate(); err != nil {
		return nil, nil, err
	}
	data, err := s.iterationData(ctx, start.Config)
	if err != nil {
		return nil, nil, err
	}
	start.Config = start.Config.WithDataIterations(data.Count())
	environment, err := s.environment(ctx, start, workspace.Path)
	if err != nil {
		return nil, nil, err
	}
	preflight, err := s.preflight(ctx, start, workspace.Path, environment.id, rows)
	if err != nil {
		return nil, nil, err
	}
	acknowledgement := start.Acknowledgement()
	if preflight.AwaitsConfirmation(acknowledgement) {
		confirmation := preflight.RunConfirmation(target)
		return nil, &confirmation, usecase.RejectUnconfirmedSend(preflight)
	}
	if preflight.AwaitsSecretsAcknowledgement(acknowledgement) {
		return nil, nil, usecase.RejectUnavailableSecrets(preflight)
	}
	run, err := s.executor.launch(ctx, start, functionalRun{target: target, rows: rows, root: workspace.Path, environment: environment, data: *data, appVersion: appVersion})
	return run, nil, err
}

func (s *runStarter) target(ctx context.Context, start entity.RunStart, root string) (entity.RunTarget, []entity.RunPlanRow, error) {
	tree, err := s.loadTree.Load(ctx, root)
	if err != nil {
		return entity.RunTarget{}, nil, err
	}
	container, err := tree.RunTarget(start.ItemID)
	if err != nil {
		return entity.RunTarget{}, nil, err
	}
	rows := start.Config.PlannedRows(entity.FlattenRunRows(*container))
	return entity.NewRunTarget(start.LocalID, *container), rows, nil
}

func (s *runStarter) environment(ctx context.Context, start entity.RunStart, root string) (runEnvironment, error) {
	local, err := s.findLocalVariables.Find(ctx, start.LocalID)
	if local, err = localVariablesOrDefaults(ctx, start.LocalID, local, err); err != nil {
		return runEnvironment{}, err
	}
	environment, err := s.pipeline.variables.environment(ctx, root, *local, start.Config.EnvironmentIn(*local))
	if err != nil || environment == nil {
		return runEnvironment{id: entity.NoEnvironmentID}, err
	}
	return runEnvironment{id: environment.ID, name: environment.Name}, nil
}

func (s *runStarter) preflight(ctx context.Context, start entity.RunStart, root, environmentId string, rows []entity.RunPlanRow) (entity.SendPreflight, error) {
	combined := entity.SendPreflight{}
	for _, row := range rows {
		stored, err := s.findRequest.Find(ctx, root, row.RequestID)
		if errs.IsNotFoundError(err) {
			continue
		}
		if err != nil {
			return entity.SendPreflight{}, err
		}
		request := stored.Version(row.AliasID)
		scopes, err := s.pipeline.variables.loadIn(ctx, start.LocalID, root, request.ID, environmentId)
		if err != nil {
			return entity.SendPreflight{}, err
		}
		effective := usecase.ResolveEffectiveRequest(request, scopes.ancestors, root)
		combined = combined.Merged(usecase.CheckSendPreflight(request.Inheriting(effective), scopes.context, scopes.environment))
	}
	return combined, nil
}

func (s *runStarter) iterationData(ctx context.Context, config entity.RunConfig) (*entity.IterationData, error) {
	if !config.HasDataFile() {
		return &entity.IterationData{}, nil
	}
	data, err := s.findIterationData.FindIterationData(ctx, config.DataFile())
	if err != nil {
		return nil, err
	}
	if err := data.Validate(); err != nil {
		return nil, err
	}
	return data, nil
}
