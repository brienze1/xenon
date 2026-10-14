package mcp

import (
	"context"

	serverAdapter "github.com/brienze1/xenon/internal/infra/server/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

type server struct {
	aiAccess adapter.AIAccessController
	listener serverAdapter.MCPListener
	routes   serverAdapter.MCPRoutes
}

func NewServer(aiAccess adapter.AIAccessController, listener serverAdapter.MCPListener, routes serverAdapter.MCPRoutes) serverAdapter.MCPServer {
	return &server{aiAccess: aiAccess, listener: listener, routes: routes}
}

func (s *server) Start(ctx context.Context) error {
	s.listener.Route(s.routes)
	s.listener.Watch(s.clientsChanged)
	_, err := s.aiAccess.Start(ctx)
	return err
}

func (s *server) Stop(ctx context.Context) error {
	return s.aiAccess.Shutdown(ctx)
}

func (s *server) clientsChanged(ctx context.Context) {
	_ = s.aiAccess.ClientsChanged(ctx)
}
