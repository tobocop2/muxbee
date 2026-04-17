// Package megabridge manages mautrix bridgev2 connectors in-process.
package megabridge

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"

	"go.mau.fi/util/dbutil"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/bridgeconfig"
	"maunium.net/go/mautrix/bridgev2/commands"
	"maunium.net/go/mautrix/bridgev2/matrix"

	"github.com/rs/zerolog"

	"github.com/tobocop2/beetrix/internal/bridges"
	"github.com/tobocop2/beetrix/internal/config"
)

// BridgeInstance holds a running bridge and its resources.
type BridgeInstance struct {
	Name   string
	Bridge *bridgev2.Bridge
	Matrix *matrix.Connector
	DB     *dbutil.Database
}

// Manager manages the lifecycle of all bridge connectors.
type Manager struct {
	cfg       *config.Config
	instances map[string]*BridgeInstance
	mu        sync.RWMutex
}

// NewManager creates a new bridge manager.
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		cfg:       cfg,
		instances: make(map[string]*BridgeInstance),
	}
}

// StartAll starts all enabled bridges that have bridgev2 support.
func (m *Manager) StartAll(ctx context.Context) error {
	for _, name := range m.cfg.EnabledBridges {
		info := bridges.Get(name)
		if info == nil || !info.BridgeV2Ready {
			continue
		}

		if err := m.startBridge(ctx, name, info); err != nil {
			slog.Error("failed to start bridge", "bridge", name, "error", err)
			continue
		}
	}
	return nil
}

// StopAll stops all running bridges.
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, inst := range m.instances {
		slog.Info("stopping bridge", "bridge", name)
		inst.Bridge.Stop()
		if inst.DB != nil {
			inst.DB.Close()
		}
		delete(m.instances, name)
	}
}

// Health returns the health status of all running bridges.
func (m *Manager) Health() map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]bool, len(m.instances))
	for name := range m.instances {
		status[name] = true
	}
	return status
}

func (m *Manager) startBridge(ctx context.Context, name string, info *bridges.BridgeInfo) error {
	slog.Info("starting bridge", "bridge", name, "port", info.Port)

	connector, err := NewConnector(name)
	if err != nil {
		return fmt.Errorf("creating connector: %w", err)
	}

	tokens, ok := m.cfg.BridgeTokens[name]
	if !ok {
		return fmt.Errorf("no tokens configured for bridge %s", name)
	}

	// Build bridge config programmatically
	bridgeCfg := buildBridgeConfig(m.cfg, name, info, tokens)

	// Create database
	log := zerolog.New(zerolog.NewConsoleWriter()).With().
		Str("bridge", name).
		Timestamp().
		Logger()

	db, err := dbutil.NewFromConfig("megabridge/"+name, bridgeCfg.Database, dbutil.ZeroLogger(log))
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}

	// Create matrix connector
	matrixConn := matrix.NewConnector(bridgeCfg)

	// Create bridge
	br := bridgev2.NewBridge("", db, log, &bridgeCfg.Bridge, matrixConn, connector, commands.NewProcessor)

	// Start bridge
	if err := br.Start(ctx); err != nil {
		db.Close()
		return fmt.Errorf("starting bridge: %w", err)
	}

	m.mu.Lock()
	m.instances[name] = &BridgeInstance{
		Name:   name,
		Bridge: br,
		Matrix: matrixConn,
		DB:     db,
	}
	m.mu.Unlock()

	slog.Info("bridge started", "bridge", name)
	return nil
}

func buildBridgeConfig(cfg *config.Config, name string, info *bridges.BridgeInfo, tokens config.BridgeTokens) *bridgeconfig.Config {
	return &bridgeconfig.Config{
		Homeserver: bridgeconfig.HomeserverConfig{
			Address:  fmt.Sprintf("http://%s:%d", cfg.Address(), cfg.Port()),
			Domain:   cfg.ServerName,
			Software: "standard",
		},
		AppService: bridgeconfig.AppserviceConfig{
			ID:       name,
			ASToken:  tokens.ASToken,
			HSToken:  tokens.HSToken,
			Bot:      bridgeconfig.BotUserConfig{Username: info.BotUsername()},
			Port:     uint16(info.Port),
			Hostname: "127.0.0.1",
			UsernameTemplate: fmt.Sprintf("%s_{{.}}", name),
		},
		Bridge: bridgeconfig.BridgeConfig{
			CommandPrefix: fmt.Sprintf("!%s", name),
			Permissions: bridgeconfig.PermissionConfig{
				"*": &bridgeconfig.Permissions{
					SendEvents: true,
					Commands:   true,
					Login:      true,
				},
			},
		},
		Database: dbutil.Config{
			PoolConfig: dbutil.PoolConfig{
				Type: "sqlite3-fk-wal",
				URI:  fmt.Sprintf("file:%s?_txlock=immediate", filepath.Join(config.BridgeDataDir(name), "bridge.db")),
			},
		},
	}
}
