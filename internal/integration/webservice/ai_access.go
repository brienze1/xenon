package webservice

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type aiAccessWebservice struct {
	listener adapter.AIAccessListener
	status   entity.AIAccessStatus
	mu       sync.Mutex
	statusMu sync.Mutex
}

func NewAIAccessWebservice(listener adapter.AIAccessListener) adapter.AIAccessWebservice {
	return &aiAccessWebservice{
		listener: listener,
		status:   entity.OffAIAccessStatus(entity.DefaultAIPort),
	}
}

func (a *aiAccessWebservice) Apply(ctx context.Context, settings entity.AIAccessSettings) entity.AIAccessStatus {
	logger.Debug(ctx, "Started", settings)

	a.mu.Lock()
	a.listener.Close()
	a.setStatus(a.listen(ctx, effectivePort(settings)))
	a.mu.Unlock()
	status := a.currentStatus()

	logger.Debug(ctx, "Finished", status)
	return status
}

func (a *aiAccessWebservice) Disable(ctx context.Context, settings entity.AIAccessSettings) entity.AIAccessStatus {
	logger.Debug(ctx, "Started", settings)

	status := a.stopAt(effectivePort(settings))

	logger.Debug(ctx, "Finished", status)
	return status
}

func (a *aiAccessWebservice) Status(ctx context.Context) entity.AIAccessStatus {
	logger.Debug(ctx, "Started")

	status := a.currentStatus()

	logger.Debug(ctx, "Finished", status)
	return status
}

func (a *aiAccessWebservice) Stop(ctx context.Context) entity.AIAccessStatus {
	logger.Debug(ctx, "Started")

	status := a.stopAt(a.currentStatus().Port)

	logger.Debug(ctx, "Finished", status)
	return status
}

func (a *aiAccessWebservice) stopAt(port int) entity.AIAccessStatus {
	a.mu.Lock()
	a.listener.Close()
	a.setStatus(entity.OffAIAccessStatus(port))
	a.mu.Unlock()
	return a.currentStatus()
}

func effectivePort(settings entity.AIAccessSettings) int {
	return settings.EffectivePort(properties.Properties().Application.MCPPortOverride)
}

func (a *aiAccessWebservice) listen(ctx context.Context, port int) entity.AIAccessStatus {
	if err := a.listener.Listen(port); err != nil {
		logger.Warn(ctx, err, "The MCP server could not listen", map[string]any{"port": port})
		return entity.PortInUseAIAccessStatus(port)
	}
	return entity.RunningAIAccessStatus(port, 0)
}

func (a *aiAccessWebservice) setStatus(status entity.AIAccessStatus) {
	a.statusMu.Lock()
	defer a.statusMu.Unlock()
	a.status = status
}

func (a *aiAccessWebservice) currentStatus() entity.AIAccessStatus {
	clients := a.listener.Clients()
	a.statusMu.Lock()
	defer a.statusMu.Unlock()
	return a.status.WithClients(clients)
}
