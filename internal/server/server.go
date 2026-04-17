// Package server orchestrates the Dendrite homeserver and bridge manager lifecycle.
package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tobocop2/beetrix/internal/config"
	"github.com/tobocop2/beetrix/internal/homeserver"
	"github.com/tobocop2/beetrix/internal/megabridge"
)

// Server composes the homeserver and bridge manager.
type Server struct {
	cfg        *config.Config
	homeserver *homeserver.DendriteServer
	bridges    *megabridge.Manager
}

// New creates a new server from config.
func New(cfg *config.Config) (*Server, error) {
	hs, err := homeserver.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating homeserver: %w", err)
	}

	return &Server{
		cfg:        cfg,
		homeserver: hs,
		bridges:    megabridge.NewManager(cfg),
	}, nil
}

// Start boots the homeserver then all enabled bridges.
func (s *Server) Start(ctx context.Context) error {
	// Start Dendrite first
	if err := s.homeserver.Start(ctx); err != nil {
		return fmt.Errorf("starting homeserver: %w", err)
	}

	// Wait for homeserver to be healthy
	if err := s.homeserver.WaitForHealth(); err != nil {
		s.homeserver.Stop()
		return fmt.Errorf("homeserver health: %w", err)
	}

	// Create admin user
	if err := s.homeserver.CreateUser(ctx, s.cfg.Admin.Username, s.cfg.Admin.Password, true); err != nil {
		slog.Warn("could not create admin user", "error", err)
	}

	// Start bridges
	if err := s.bridges.StartAll(ctx); err != nil {
		slog.Error("some bridges failed to start", "error", err)
	}

	return nil
}

// Stop shuts down bridges first, then the homeserver.
func (s *Server) Stop() {
	slog.Info("stopping server")
	s.bridges.StopAll()
	s.homeserver.Stop()
}

// Health returns aggregated health status.
func (s *Server) Health() Health {
	return Health{
		Homeserver: s.homeserver.Health() == nil,
		Bridges:    s.bridges.Health(),
	}
}

// Health holds the health status of all components.
type Health struct {
	Homeserver bool
	Bridges    map[string]bool
}
