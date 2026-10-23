package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type runEnvironment struct {
	id   string
	name string
}

func (m *manageRunner) target(ctx context.Context, start entity.RunStart, root string) (entity.RunTarget, []entity.RunPlanRow, error) {
	tree, err := m.loadTree.Load(ctx, root)
	if err != nil {
		return entity.RunTarget{}, nil, err
	}
	container, found := tree.Container(start.ItemID)
	if !found {
		return entity.RunTarget{}, nil, usecase.RejectUnknownRunTarget(start.ItemID)
	}
	sidebar := entity.FlattenRunRows(*container)
	rows := make([]entity.RunPlanRow, 0, len(start.Config.Rows))
	for _, row := range start.Config.SelectedRows() {
		planned, _ := entity.RunPlanRowOf(sidebar, row.RequestID)
		planned = planned.For(row)
		planned.Number = len(rows) + 1
		rows = append(rows, planned)
	}
	return entity.NewRunTarget(start.LocalID, *container), rows, nil
}

func (m *manageRunner) environment(ctx context.Context, start entity.RunStart, root string) (runEnvironment, error) {
	local, err := m.findLocalVariables.Find(ctx, start.LocalID)
	if local, err = localVariablesOrDefaults(ctx, start.LocalID, local, err); err != nil {
		return runEnvironment{}, err
	}
	id := start.Config.EnvironmentID
	if id == "" && local.HasActiveEnvironment() {
		id = local.ActiveEnvironmentID
	}
	if id == "" {
		id = entity.NoEnvironmentID
	}
	environment, err := m.core.variables.environment(ctx, root, *local, id)
	if err != nil || environment == nil {
		return runEnvironment{id: entity.NoEnvironmentID}, err
	}
	return runEnvironment{id: environment.ID, name: environment.Name}, nil
}

func (m *manageRunner) preflight(ctx context.Context, start entity.RunStart, root, environmentId string, rows []entity.RunPlanRow) error {
	combined := entity.SendPreflight{}
	for _, row := range rows {
		stored, err := m.findRequest.Find(ctx, root, row.RequestID)
		if err != nil {
			continue
		}
		request := stored.Version(row.AliasID)
		scopes, err := m.core.variables.loadIn(ctx, start.LocalID, root, request.ID, environmentId)
		if err != nil {
			return err
		}
		effective := usecase.ResolveEffectiveRequest(request, scopes.ancestors, root)
		combined = combined.Merged(usecase.CheckSendPreflight(request.Inheriting(effective), scopes.context, scopes.environment))
	}
	acknowledgement := start.Acknowledgement()
	if combined.AwaitsConfirmation(acknowledgement) {
		return usecase.RejectUnconfirmedSend(combined)
	}
	if combined.AwaitsSecretsAcknowledgement(acknowledgement) {
		return usecase.RejectUnavailableSecrets(combined)
	}
	return nil
}

func (m *manageRunner) launch(ctx context.Context, start entity.RunStart, run functionalRun) (*entity.Run, error) {
	run.id = idutils.NewUlid()
	run.config = start.Config
	run.acknowledgement = start.Acknowledgement()
	run.jarId = entity.RunCookieJarID(start.LocalID, run.id)
	if !start.Config.Functional.KeepVariableValues {
		run.variables = entity.NewRunVariables()
	}
	started := entity.NewRun(run.id, run.target, start.Config, len(run.rows)*start.Config.Functional.Iterations, timeutils.Now())
	started.EnvironmentName = run.environment.name
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	control := entity.NewRunControl(cancel)
	if err := m.manageActiveRuns.Claim(ctx, started, control); err != nil {
		cancel()
		return nil, err
	}
	if err := m.seedJar(ctx, start, run.jarId); err != nil {
		m.manageActiveRuns.Release(ctx, start.LocalID, run.id)
		cancel()
		return nil, err
	}
	run.control = control
	go m.runFunctional(runCtx, run)
	return &started, nil
}

func (m *manageRunner) seedJar(ctx context.Context, start entity.RunStart, jarId string) error {
	if start.Config.Functional.SkipStoredCookies {
		return nil
	}
	if err := m.core.cookies.load(ctx, start.LocalID); err != nil {
		return err
	}
	m.manageCookieJar.LoadCookies(ctx, jarId, m.manageCookieJar.JarState(ctx, start.LocalID).Cookies)
	return nil
}
