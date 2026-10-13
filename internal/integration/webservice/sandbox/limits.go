package sandbox

import (
	"runtime/metrics"
	"sync"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/scripterrorkind"
)

const (
	maxCallStackSize   = 10_000
	memoryGrowthLimit  = 512 << 20
	memoryPollInterval = 50 * time.Millisecond
	interruptInterval  = 5 * time.Millisecond
	heapObjectsMetric  = "/memory/classes/heap/objects:bytes"
)

type limits struct {
	timeout      time.Duration
	pollInterval time.Duration
	memoryGrowth uint64
}

func newLimits(settings entity.AppSettings) limits {
	return limits{timeout: settings.ScriptTimeout(), pollInterval: memoryPollInterval, memoryGrowth: memoryGrowthLimit}
}

type stopper interface {
	Interrupt(v any)
	StopNoWait()
}

type guard struct {
	done    chan struct{}
	stopped chan struct{}
	tripped enums.ScriptErrorKind
	mutex   sync.Mutex
}

func (l limits) watch(target stopper) *guard {
	g := &guard{done: make(chan struct{}), stopped: make(chan struct{})}
	go g.run(l, target)
	return g
}

func (g *guard) run(l limits, target stopper) {
	defer close(g.stopped)
	baseline := heapBytes()
	deadline := time.NewTimer(l.timeout)
	defer deadline.Stop()
	poll := time.NewTicker(l.pollInterval)
	defer poll.Stop()
	for {
		select {
		case <-g.done:
			return
		case <-deadline.C:
			g.trip(scripterrorkind.Timeout, target)
			return
		case <-poll.C:
			if exceedsGrowth(heapBytes(), baseline, l.memoryGrowth) {
				g.trip(scripterrorkind.Memory, target)
				return
			}
		}
	}
}

func (g *guard) trip(kind enums.ScriptErrorKind, target stopper) {
	g.mutex.Lock()
	g.tripped = kind
	g.mutex.Unlock()
	interrupt := time.NewTicker(interruptInterval)
	defer interrupt.Stop()
	for {
		target.Interrupt(kind)
		target.StopNoWait()
		select {
		case <-g.done:
			return
		case <-interrupt.C:
		}
	}
}

func (g *guard) stop() enums.ScriptErrorKind {
	close(g.done)
	<-g.stopped
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return g.tripped
}

func exceedsGrowth(current, baseline, growth uint64) bool {
	return current > baseline+growth
}

func heapBytes() uint64 {
	sample := []metrics.Sample{{Name: heapObjectsMetric}}
	metrics.Read(sample)
	return sample[0].Value.Uint64()
}
