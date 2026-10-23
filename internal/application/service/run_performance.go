package service

import (
	"context"
	"strconv"
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
	started   int
}

func (m *manageRunner) runPerformance(ctx context.Context, run functionalRun) {
	defer run.control.Finished()
	logger.Info(ctx, "Started", map[string]any{"localId": run.target.LocalID, "itemId": run.target.ItemID, "runId": run.id})

	performance := &performanceRun{base: run, stats: entity.NewPerformanceStats(run.rows), startedAt: time.Now()}
	status := m.drive(ctx, performance)
	m.settle(context.WithoutCancel(ctx), performance, status)
	m.finish(context.WithoutCancel(ctx), run, status)

	logger.Info(ctx, "Finished", map[string]any{"localId": run.target.LocalID, "itemId": run.target.ItemID, "runId": run.id, "status": status})
}

func (m *manageRunner) drive(ctx context.Context, run *performanceRun) enums.RunStatus {
	settings := run.base.config.Performance
	ticker := time.NewTicker(performanceTick)
	defer ticker.Stop()
	deadline := time.NewTimer(settings.Duration())
	defer deadline.Stop()
	reported := 1
	for {
		elapsed := time.Since(run.startedAt)
		if elapsed >= settings.Duration() {
			return runstatus.Completed
		}
		active := settings.ActiveUsersAt(elapsed)
		m.staff(ctx, run, active)
		second := int(elapsed / time.Second)
		run.stats.Observe(second, active)
		if second >= reported {
			m.measure(ctx, run, active, elapsed, second)
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

func (m *manageRunner) staff(ctx context.Context, run *performanceRun, active int) {
	for len(run.users) < active {
		m.startUser(ctx, run)
	}
	for len(run.users) > active {
		last := len(run.users) - 1
		close(run.users[last].retired)
		run.users = run.users[:last]
	}
}

func (m *manageRunner) startUser(ctx context.Context, run *performanceRun) {
	run.started++
	user := &virtualUser{retired: make(chan struct{}), run: run.base}
	user.run.jarId = run.base.jarId + ":" + strconv.Itoa(run.started)
	user.run.variables = entity.NewRunVariables()
	user.run.secrets = nil
	m.manageCookieJar.LoadCookies(ctx, user.run.jarId, m.manageCookieJar.JarState(ctx, run.base.jarId).Cookies)
	run.jars = append(run.jars, user.run.jarId)
	run.users = append(run.users, user)
	run.waiting.Add(1)
	go m.loopUser(ctx, run, user)
}

func (m *manageRunner) loopUser(ctx context.Context, run *performanceRun, user *virtualUser) {
	defer run.waiting.Done()
	for pass := 0; ; pass++ {
		for index, row := range run.base.rows {
			if user.isRetired() || ctx.Err() != nil {
				return
			}
			ran := m.runRow(ctx, &user.run, 0, pass, row)
			if ctx.Err() != nil {
				return
			}
			m.observe(ctx, run, index, ran)
		}
	}
}

func (m *manageRunner) observe(ctx context.Context, run *performanceRun, index int, ran ranRow) {
	if ran.next != nil {
		run.noticed.Do(func() {
			m.change(ctx, run.base, func(active *entity.Run) { active.Notice(entity.PerformanceNextRequestNotice) })
		})
	}
	if ran.result.IsMeasured() {
		run.stats.Record(entity.NewPerformanceSample(ran.result, ran.latency, index, run.second()))
	}
}

func (m *manageRunner) settle(ctx context.Context, run *performanceRun, status enums.RunStatus) {
	m.staff(ctx, run, 0)
	run.waiting.Wait()
	for _, jar := range run.jars {
		m.manageCookieJar.DropJar(ctx, jar)
	}
	elapsed := time.Since(run.startedAt)
	seconds := run.base.config.Performance.DurationSeconds
	if status != runstatus.Completed {
		seconds = min(seconds, int((elapsed+time.Second-1)/time.Second))
	}
	m.measure(ctx, run, 0, elapsed, seconds)
}

func (m *manageRunner) measure(ctx context.Context, run *performanceRun, active int, elapsed time.Duration, seconds int) {
	metrics := run.stats.Metrics(active, elapsed)
	report := run.stats.Report(elapsed, seconds)
	if measured, found := m.change(ctx, run.base, func(active *entity.Run) { active.Measure(report, metrics, timeutils.Now()) }); found {
		m.publishRunEvent.Progress(ctx, measured.Run)
	}
}

func (r *performanceRun) second() int {
	return min(int(time.Since(r.startedAt)/time.Second), r.base.config.Performance.DurationSeconds-1)
}

func (v *virtualUser) isRetired() bool {
	select {
	case <-v.retired:
		return true
	default:
		return false
	}
}
