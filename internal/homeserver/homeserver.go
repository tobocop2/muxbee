// Package homeserver embeds a Dendrite Matrix homeserver in-process.
package homeserver

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	dendembed "github.com/element-hq/dendrite/pkg/embed"

	"github.com/tobocop2/beetrix/internal/config"
)

// DendriteServer wraps an in-process Dendrite monolith.
type DendriteServer struct {
	cfg    *config.Config
	server *dendembed.Server
}

// New creates a new DendriteServer from beetrix config.
func New(cfg *config.Config) (*DendriteServer, error) {
	return &DendriteServer{cfg: cfg}, nil
}

// Start boots the Dendrite monolith and begins serving HTTP.
func (d *DendriteServer) Start(ctx context.Context) error {
	slog.Info("starting dendrite homeserver",
		"server_name", d.cfg.ServerName,
		"address", d.cfg.Address(),
		"port", d.cfg.Port(),
	)

	sk, err := loadOrGenerateSigningKey()
	if err != nil {
		return fmt.Errorf("signing key: %w", err)
	}

	dendCfg, err := dendembed.BuildConfig(dendembed.ConfigOptions{
		ServerName: d.cfg.ServerName,
		DataDir:    config.DendriteDir(),
		PrivateKey: sk,
		Federation: d.cfg.Federation,
	})
	if err != nil {
		return fmt.Errorf("building dendrite config: %w", err)
	}

	// Register appservices for enabled bridges BEFORE starting the monolith.
	// Dendrite must know about appservices at boot time for token validation.
	for name := range d.cfg.BridgeTokens {
		if !d.cfg.IsBridgeEnabled(name) {
			continue
		}
		tokens := d.cfg.BridgeTokens[name]
		dendembed.AddAppservice(dendCfg, dendembed.AppserviceRegistration{
			ID:              name,
			URL:             fmt.Sprintf("http://127.0.0.1:%d", bridgePort(name)),
			ASToken:         tokens.ASToken,
			HSToken:         tokens.HSToken,
			BotUsername:     name + "bot",
			NamespacePrefix: name + "_",
			ServerName:      d.cfg.ServerName,
		})
		slog.Info("registered appservice", "bridge", name)
	}

	listenAddr := fmt.Sprintf("%s:%d", d.cfg.Address(), d.cfg.Port())
	d.server, err = dendembed.Start(dendCfg, listenAddr)
	if err != nil {
		return fmt.Errorf("starting dendrite: %w", err)
	}

	// Propagate parent context cancellation
	go func() {
		<-ctx.Done()
		d.Stop()
	}()

	slog.Info("dendrite homeserver started", "addr", listenAddr)
	return nil
}

// Stop gracefully shuts down the Dendrite server.
func (d *DendriteServer) Stop() {
	if d.server != nil {
		slog.Info("stopping dendrite homeserver")
		d.server.Stop()
		slog.Info("dendrite homeserver stopped")
	}
}

// WaitForHealth polls until the homeserver responds or times out.
func (d *DendriteServer) WaitForHealth() error {
	for i := 0; i < 30; i++ {
		if err := d.Health(); err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("homeserver did not become healthy within 15 seconds")
}

// Health checks if the homeserver is responding.
func (d *DendriteServer) Health() error {
	url := fmt.Sprintf("http://%s:%d/_matrix/client/versions", d.cfg.Address(), d.cfg.Port())
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}
	return nil
}

// CreateUser creates a Matrix user account.
func (d *DendriteServer) CreateUser(ctx context.Context, username, password string, admin bool) error {
	if d.server == nil {
		return fmt.Errorf("homeserver not started")
	}
	return d.server.CreateUser(ctx, username, password, admin)
}

// bridgePort returns the port for a named bridge from the registry.
// Falls back to 0 if bridge is unknown.
func bridgePort(name string) int {
	// Default ports per bridge (matches bridges.yaml)
	ports := map[string]int{
		"whatsapp": 29318, "signal": 29313, "slack": 29315,
		"meta": 29323, "gmessages": 29314, "twitter": 29324,
		"bluesky": 29325, "linkedin": 29322, "irc": 29326,
		"googlechat": 29320, "gvoice": 29321, "discord": 29316,
		"telegram": 29317,
	}
	if p, ok := ports[name]; ok {
		return p
	}
	return 0
}

// loadOrGenerateSigningKey loads the server signing key from disk, or generates a new one.
func loadOrGenerateSigningKey() (ed25519.PrivateKey, error) {
	keyPath := filepath.Join(config.DendriteDir(), "signing_key.pem")

	if data, err := os.ReadFile(keyPath); err == nil {
		if len(data) == ed25519.SeedSize {
			return ed25519.NewKeyFromSeed(data), nil
		}
		if len(data) == ed25519.PrivateKeySize {
			return ed25519.PrivateKey(data), nil
		}
	}

	_, sk, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(keyPath, sk.Seed(), 0600); err != nil {
		return nil, fmt.Errorf("saving signing key: %w", err)
	}

	return sk, nil
}
