package mcp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	serverName        = "xenon"
	endpointPath      = "/mcp"
	readHeaderTimeout = 10 * time.Second
)

var legacyProtocolVersions = []string{"2025-11-25", "2025-06-18", "2025-03-26", "2024-11-05"}

// SessionClosed is called with the id of every MCP session that ends, whatever ended it.
type SessionClosed func(ctx context.Context, sessionId string)

type Server struct {
	publisher   adapter.AIEventPublisher
	tools       *tools
	httpServer  *http.Server
	sdkServer   *sdk.Server
	sessions    map[string]string
	closeHooks  []SessionClosed
	status      entity.AIAccessStatus
	published   entity.AIAccessStatus
	mu          sync.Mutex
	statusMu    sync.Mutex
	sessionsMu  sync.Mutex
	hooksMu     sync.Mutex
	publishedMu sync.Mutex
}

func NewServer(publisher adapter.AIEventPublisher) *Server {
	return &Server{
		publisher: publisher,
		sessions:  map[string]string{},
		status:    entity.OffAIAccessStatus(entity.DefaultAIPort),
		published: entity.OffAIAccessStatus(entity.DefaultAIPort),
	}
}

func (s *Server) Mount(controllers Controllers) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools = newTools(controllers)
}

func (s *Server) OnSessionClosed(hook SessionClosed) {
	s.hooksMu.Lock()
	defer s.hooksMu.Unlock()
	s.closeHooks = append(s.closeHooks, hook)
}

func (s *Server) Apply(ctx context.Context, settings entity.AIAccessSettings) entity.AIAccessStatus {
	logger.Debug(ctx, "Started", settings)

	s.mu.Lock()
	s.stop()
	port := settings.EffectivePort(properties.Properties().Application.MCPPortOverride)
	applied := entity.OffAIAccessStatus(port)
	if settings.Enabled {
		applied = s.start(ctx, port)
	}
	s.setStatus(applied)
	s.mu.Unlock()
	status := s.currentStatus()
	s.publish(ctx, status)

	logger.Debug(ctx, "Finished", status)
	return status
}

func (s *Server) Status(ctx context.Context) entity.AIAccessStatus {
	logger.Debug(ctx, "Started")

	status := s.currentStatus()

	logger.Debug(ctx, "Finished", status)
	return status
}

func (s *Server) Stop(ctx context.Context) entity.AIAccessStatus {
	logger.Debug(ctx, "Started")

	s.mu.Lock()
	s.stop()
	s.setStatus(entity.OffAIAccessStatus(s.currentStatus().Port))
	s.mu.Unlock()
	status := s.currentStatus()
	s.publish(ctx, status)

	logger.Debug(ctx, "Finished", status)
	return status
}

func (s *Server) start(ctx context.Context, port int) entity.AIAccessStatus {
	listener, err := net.Listen("tcp", entity.AIAccessListenAddress(port))
	if err != nil {
		logger.Warn(ctx, err, "The MCP server could not listen", map[string]any{"port": port})
		return entity.PortInUseAIAccessStatus(port)
	}
	s.sdkServer = s.newSdkServer()
	handler := sdk.NewStreamableHTTPHandler(func(*http.Request) *sdk.Server { return s.sdkServer }, &sdk.StreamableHTTPOptions{})
	mux := http.NewServeMux()
	mux.Handle(endpointPath, handler)
	s.httpServer = &http.Server{Handler: loopbackOnly(port, mux), ReadHeaderTimeout: readHeaderTimeout}
	go s.serve(s.httpServer, listener)
	return entity.RunningAIAccessStatus(port, 0)
}

func (s *Server) serve(server *http.Server, listener net.Listener) {
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Warn(context.Background(), err, "The MCP server stopped")
	}
}

func (s *Server) stop() {
	if s.sdkServer != nil {
		for session := range s.sdkServer.Sessions() {
			_ = session.Close()
		}
	}
	if s.httpServer != nil {
		_ = s.httpServer.Close()
	}
	s.httpServer, s.sdkServer = nil, nil
}

func (s *Server) newSdkServer() *sdk.Server {
	server := sdk.NewServer(&sdk.Implementation{Name: serverName, Version: properties.Properties().Application.Version}, &sdk.ServerOptions{
		InitializedHandler:        s.track,
		SupportedProtocolVersions: legacyProtocolVersions,
		Capabilities:              &sdk.ServerCapabilities{Tools: &sdk.ToolCapabilities{}},
	})
	for _, definition := range registry {
		server.AddTool(definition.definition(), pipeline{tools: s.tools, tool: definition}.handle)
	}
	return server
}

func (s *Server) track(ctx context.Context, request *sdk.InitializedRequest) {
	session := request.Session
	if session == nil {
		return
	}
	client := entity.DefaultAIClientName
	if params := session.InitializeParams(); params != nil && params.ClientInfo != nil {
		client = entity.AIClientName(params.ClientInfo.Name)
	}
	s.sessionsMu.Lock()
	s.sessions[session.ID()] = client
	s.sessionsMu.Unlock()
	s.publish(ctx, s.currentStatus())
	go s.untrackWhenClosed(session)
}

func (s *Server) untrackWhenClosed(session *sdk.ServerSession) {
	_ = session.Wait()
	ctx := context.Background()
	s.sessionsMu.Lock()
	delete(s.sessions, session.ID())
	s.sessionsMu.Unlock()
	s.publish(ctx, s.currentStatus())
	for _, hook := range s.hooks() {
		hook(ctx, session.ID())
	}
}

func (s *Server) hooks() []SessionClosed {
	s.hooksMu.Lock()
	defer s.hooksMu.Unlock()
	return append([]SessionClosed{}, s.closeHooks...)
}

func (s *Server) setStatus(status entity.AIAccessStatus) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	s.status = status
}

func (s *Server) currentStatus() entity.AIAccessStatus {
	s.sessionsMu.Lock()
	clients := len(s.sessions)
	s.sessionsMu.Unlock()
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	return s.status.WithClients(clients)
}

func (s *Server) publish(ctx context.Context, status entity.AIAccessStatus) {
	s.publishedMu.Lock()
	defer s.publishedMu.Unlock()
	if status == s.published {
		return
	}
	s.published = status
	if err := s.publisher.AccessStatusChanged(ctx, status); err != nil {
		logger.Warn(ctx, err, "The AI access status could not be published", status)
	}
}
