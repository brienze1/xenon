package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type runExecutor struct {
	pipeline         *sendPipeline
	findRequest      usecase.FindRequest
	manageCookieJar  usecase.ManageCookieJar
	claimActiveRun   usecase.ClaimActiveRun
	changeActiveRun  usecase.ChangeActiveRun
	releaseActiveRun usecase.ReleaseActiveRun
	findActiveRuns   usecase.FindActiveRuns
	saveRun          usecase.SaveRun
	findRuns         usecase.FindRuns
	deleteRuns       usecase.DeleteRuns
	publishRunEvent  usecase.PublishRunEvent
	releaseResponse  usecase.ReleaseResponse
}

func newRunExecutor(
	pipeline *sendPipeline,
	findRequest usecase.FindRequest,
	manageCookieJar usecase.ManageCookieJar,
	claimActiveRun usecase.ClaimActiveRun,
	changeActiveRun usecase.ChangeActiveRun,
	releaseActiveRun usecase.ReleaseActiveRun,
	findActiveRuns usecase.FindActiveRuns,
	saveRun usecase.SaveRun,
	findRuns usecase.FindRuns,
	deleteRuns usecase.DeleteRuns,
	publishRunEvent usecase.PublishRunEvent,
	releaseResponse usecase.ReleaseResponse,
) *runExecutor {
	return &runExecutor{
		pipeline:         pipeline,
		findRequest:      findRequest,
		manageCookieJar:  manageCookieJar,
		claimActiveRun:   claimActiveRun,
		changeActiveRun:  changeActiveRun,
		releaseActiveRun: releaseActiveRun,
		findActiveRuns:   findActiveRuns,
		saveRun:          saveRun,
		findRuns:         findRuns,
		deleteRuns:       deleteRuns,
		publishRunEvent:  publishRunEvent,
		releaseResponse:  releaseResponse,
	}
}

func (e *runExecutor) launch(ctx context.Context, start entity.RunStart, run functionalRun) (*entity.Run, error) {
	run.id = idutils.NewUlid()
	run.config = start.Config
	run.acknowledgement = start.Acknowledgement()
	run.jarId = entity.RunCookieJarID(start.LocalID, run.id)
	if !start.Config.KeepsVariableValues() {
		run.variables = entity.NewRunVariables()
	}
	started := entity.NewRun(run.id, run.target, start.Config, start.Config.PlannedRequests(len(run.rows)), timeutils.Now())
	started.RecordSource(run.environment.name, run.data.FileName)
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	control := newRunControl(cancel)
	if err := e.claimActiveRun.Claim(ctx, started, control); err != nil {
		cancel()
		return nil, err
	}
	if err := e.seedJar(ctx, start, run.jarId); err != nil {
		e.releaseActiveRun.Release(ctx, start.LocalID, run.id)
		cancel()
		return nil, err
	}
	run.control = control
	if start.Config.IsPerformance() {
		go e.runPerformance(runCtx, run)
		return &started, nil
	}
	go e.runFunctional(runCtx, run)
	return &started, nil
}

func (e *runExecutor) seedJar(ctx context.Context, start entity.RunStart, jarId string) error {
	if !start.Config.SeedsStoredCookies() {
		return nil
	}
	if err := e.pipeline.cookies.load(ctx, start.LocalID); err != nil {
		return err
	}
	e.manageCookieJar.LoadCookies(ctx, jarId, e.manageCookieJar.JarState(ctx, start.LocalID).Cookies)
	return nil
}
