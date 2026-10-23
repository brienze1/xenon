package service

import (
	"context"
	"sync"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/runstatus"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const performanceTick = 100 * time.Millisecond

type virtualUser struct {
	retired chan struct{}
	run     functionalRun
}

type performanceRun struct {
	startedAt time.Time
	stats     *entity.PerformanceStats
	users     []*virtualUser
	jars      []string
	base      functionalRun
	waiting   sync.WaitGroup
	noticed   sync.Once
	mutex     sync.Mutex
	started   int
}

func (e *runExecutor) runPerformance(ctx context.Context, run functionalRun) {
	defer run.control.finished()
	logger.Info(ctx, "Started", map[string]any{"localId": run.target.LocalID, "itemId": run.target.ItemID, "runId": run.id})

	performance := &performanceRun{base: run, stats: entity.NewPerformanceStats(run.rows), startedAt: timeutils.Now()}
	status := e.drive(ctx, performance)
	e.settle(context.WithoutCancel(ctx), performance, status)
	e.finish(context.WithoutCancel(ctx), run, status)

	logger.Info(ctx, "Finished", map[string]any{"localId": run.target.LocalID, "itemId": run.target.ItemID, "runId": run.id, "status": status})
}

func (e *runExecutor) drive(ctx context.Context, run *performanceRun) enums.RunStatus {
	settings := run.base.config.Performance
	ticker := time.NewTicker(performanceTick)
	defer ticker.Stop()
	deadline := time.NewTimer(settings.Duration())
	defer deadline.Stop()
	reported := 1
	for {
		elapsed := run.elapsed()
		if settings.HasElapsed(elapsed) {
			return runstatus.Completed
		}
		active := settings.ActiveUsersAt(elapsed)
		e.staff(ctx, run, active)
		second := settings.SecondAt(elapsed)
		run.observe(second, active)
		if second >= reported {
			e.measure(ctx, run, active, elapsed, second)
			reported = second + 1
		}
		select {
		case <-ticker.C:
		case <-deadline.C:
			return runstatus.Completed
		case <-ctx.Done():
			return runstatus.Stopped
		}
	}
}

func (e *runExecutor) staff(ctx context.Context, run *performanceRun, active int) {
	for len(run.users) < active {
		e.startUser(ctx, run)
	}
	for len(run.users) > active {
		last := len(run.users) - 1
		close(run.users[last].retired)
		run.users = run.users[:last]
	}
}

func (e *runExecutor) startUser(ctx context.Context, run *performanceRun) {
	run.started++
	user := &virtualUser{retired: make(chan struct{}), run: run.base}
	user.run.jarId = entity.VirtualUserCookieJarID(run.base.jarId, run.started)
	user.run.variables = entity.NewRunVariables()
	user.run.secrets = nil
	e.manageCookieJar.LoadCookies(ctx, user.run.jarId, e.manageCookieJar.JarState(ctx, run.base.jarId).Cookies)
	run.jars = append(run.jars, user.run.jarId)
	run.users = append(run.users, user)
	run.waiting.Add(1)
	go e.loopUser(ctx, run, user)
}

func (e *runExecutor) loopUser(ctx context.Context, run *performanceRun, user *virtualUser) {
	defer run.waiting.Done()
	for pass := 0; ; pass++ {
		for index, row := range run.base.rows {
			if user.isRetired() || ctx.Err() != nil {
				return
			}
			ran := e.runRow(ctx, &user.run, 0, pass, row)
			if ctx.Err() != nil {
				return
			}
			e.observe(ctx, run, index, ran)
		}
	}
}

func (e *runExecutor) observe(ctx context.Context, run *performanceRun, index int, ran ranRow) {
	if ran.next != nil {
		run.noticed.Do(func() {
			e.change(ctx, run.base, func(active *entity.Run) { active.Notice(entity.PerformanceNextRequestNotice) })
		})
	}
	if ran.result.IsMeasured() {
		run.record(entity.NewPerformanceSample(ran.result, ran.latency, index, run.second()))
	}
}

func (e *runExecutor) settle(ctx context.Context, run *performanceRun, status enums.RunStatus) {
	e.staff(ctx, run, 0)
	run.waiting.Wait()
	for _, jar := range run.jars {
		e.manageCookieJar.DropJar(ctx, jar)
	}
	elapsed := run.elapsed()
	e.measure(ctx, run, 0, elapsed, run.base.config.Performance.ReportedSeconds(status, elapsed))
}

func (e *runExecutor) measure(ctx context.Context, run *performanceRun, active int, elapsed time.Duration, seconds int) {
	metrics, report := run.figures(active, elapsed, seconds)
	if measured, found := e.change(ctx, run.base, func(active *entity.Run) { active.Measure(report, metrics, timeutils.Now()) }); found {
		if err := e.publishRunEvent.Progress(ctx, *measured); err != nil {
			logger.Warn(ctx, err, "The run progress could not be published", map[string]any{"runId": run.base.id})
		}
	}
}

func (r *performanceRun) second() int {
	return r.base.config.Performance.SecondAt(r.elapsed())
}

func (r *performanceRun) elapsed() time.Duration {
	return timeutils.Now().Sub(r.startedAt)
}

func (r *performanceRun) record(sample entity.PerformanceSample) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.stats.Record(sample)
}

func (r *performanceRun) observe(second, activeUsers int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.stats.Observe(second, activeUsers)
}

func (r *performanceRun) figures(activeUsers int, elapsed time.Duration, seconds int) (entity.PerformanceMetrics, entity.PerformanceReport) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.stats.Metrics(activeUsers, elapsed), r.stats.Report(elapsed, seconds)
}

func (v *virtualUser) isRetired() bool {
	select {
	case <-v.retired:
		return true
	default:
		return false
	}
}
