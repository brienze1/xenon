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
	serverAdapter "github.com/brienze1/xenon/internal/infra/server/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	serverName        = "xenon"
	readHeaderTimeout = 10 * time.Second
)

var legacyProtocolVersions = []string{"2025-11-25", "2025-06-18", "2025-03-26", "2024-11-05"}

type listener struct {
	routes         serverAdapter.MCPRoutes
	httpServer     *http.Server
	netListener    net.Listener
	sdkServer      *sdk.Server
	sessions       map[string]struct{}
	clientsChanged func(ctx context.Context)
	mu             sync.Mutex
	sessionsMu     sync.Mutex
	notifyMu       sync.Mutex
}

func NewListener() serverAdapter.MCPListener {
	return &listener{sessions: map[string]struct{}{}}
}

func (l *listener) Route(routes serverAdapter.MCPRoutes) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.routes = routes
}

func (l *listener) Watch(clientsChanged func(ctx context.Context)) {
	l.notifyMu.Lock()
	defer l.notifyMu.Unlock()
	l.clientsChanged = clientsChanged
}

func (l *listener) Listen(port int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	netListener, err := net.Listen("tcp", entity.AIAccessListenAddress(port))
	if err != nil {
		return err
	}
	sdkServer := l.newSdkServer()
	endpoint := sdk.NewStreamableHTTPHandler(func(*http.Request) *sdk.Server { return sdkServer }, &sdk.StreamableHTTPOptions{})
	l.sdkServer = sdkServer
	l.httpServer = &http.Server{Handler: l.routes.Handler(endpoint), ReadHeaderTimeout: readHeaderTimeout}
	l.netListener = netListener
	go serve(l.httpServer, netListener)
	return nil
}

func (l *listener) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.sdkServer != nil {
		for session := range l.sdkServer.Sessions() {
			_ = session.Close()
		}
	}
	if l.httpServer != nil {
		_ = l.httpServer.Close()
	}
	// Serve runs on its own goroutine, so http.Server may not hold the socket yet: close it here to free the port at once
	if l.netListener != nil {
		_ = l.netListener.Close()
	}
	l.httpServer, l.sdkServer, l.netListener = nil, nil, nil
}

func (l *listener) Clients() int {
	l.sessionsMu.Lock()
	defer l.sessionsMu.Unlock()
	return len(l.sessions)
}

func serve(server *http.Server, netListener net.Listener) {
	if err := server.Serve(netListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Warn(context.Background(), err, "The MCP server stopped")
	}
}

func (l *listener) newSdkServer() *sdk.Server {
	server := sdk.NewServer(&sdk.Implementation{Name: serverName, Version: properties.Properties().Application.Version}, &sdk.ServerOptions{
		InitializedHandler:        l.track,
		SupportedProtocolVersions: legacyProtocolVersions,
		Capabilities:              &sdk.ServerCapabilities{Tools: &sdk.ToolCapabilities{}},
	})
	l.routes.Register(server)
	return server
}

func (l *listener) track(ctx context.Context, request *sdk.InitializedRequest) {
	session := request.Session
	if session == nil {
		return
	}
	l.sessionsMu.Lock()
	l.sessions[session.ID()] = struct{}{}
	l.sessionsMu.Unlock()
	l.notifyClientsChanged(ctx)
	go l.untrackWhenClosed(session)
}

func (l *listener) untrackWhenClosed(session *sdk.ServerSession) {
	_ = session.Wait()
	ctx := context.Background()
	l.sessionsMu.Lock()
	delete(l.sessions, session.ID())
	l.sessionsMu.Unlock()
	l.notifyClientsChanged(ctx)
	l.routes.CloseSession(ctx, session.ID())
}

func (l *listener) notifyClientsChanged(ctx context.Context) {
	l.notifyMu.Lock()
	clientsChanged := l.clientsChanged
	l.notifyMu.Unlock()
	if clientsChanged != nil {
		clientsChanged(ctx)
	}
}
