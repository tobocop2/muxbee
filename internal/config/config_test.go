package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefault(t *testing.T) {
	cfg, err := NewDefault()
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.ServerName)
	assert.Equal(t, "127.0.0.1", cfg.ListenAddress)
	assert.Equal(t, 8008, cfg.ListenPort)
	assert.False(t, cfg.Federation)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "admin", cfg.Admin.Username)
	assert.NotEmpty(t, cfg.Admin.Password)
	assert.Empty(t, cfg.EnabledBridges)
}

func TestConfig_Port(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		expected int
	}{
		{"default", 0, 8008},
		{"custom", 9090, 9090},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{ListenPort: tt.port}
			assert.Equal(t, tt.expected, cfg.Port())
		})
	}
}

func TestConfig_Address(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected string
	}{
		{"default", "", "127.0.0.1"},
		{"custom", "0.0.0.0", "0.0.0.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{ListenAddress: tt.address}
			assert.Equal(t, tt.expected, cfg.Address())
		})
	}
}

func TestConfig_BridgeEnableDisable(t *testing.T) {
	cfg := &Config{EnabledBridges: []string{}}

	assert.False(t, cfg.IsBridgeEnabled("whatsapp"))

	cfg.EnableBridge("whatsapp")
	assert.True(t, cfg.IsBridgeEnabled("whatsapp"))

	// Idempotent
	cfg.EnableBridge("whatsapp")
	assert.Len(t, cfg.EnabledBridges, 1)

	cfg.DisableBridge("whatsapp")
	assert.False(t, cfg.IsBridgeEnabled("whatsapp"))

	// Idempotent
	cfg.DisableBridge("whatsapp")
	assert.Empty(t, cfg.EnabledBridges)
}

func TestConfig_GetOrCreateBridgeTokens(t *testing.T) {
	cfg := &Config{}

	tokens, err := cfg.GetOrCreateBridgeTokens("whatsapp")
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.ASToken)
	assert.NotEmpty(t, tokens.HSToken)

	// Returns same tokens on second call
	tokens2, err := cfg.GetOrCreateBridgeTokens("whatsapp")
	require.NoError(t, err)
	assert.Equal(t, tokens, tokens2)
}

func TestConfig_GetOrCreateDoublePuppetTokens(t *testing.T) {
	cfg := &Config{}

	tokens, err := cfg.GetOrCreateDoublePuppetTokens()
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.ASToken)
	assert.NotEmpty(t, tokens.HSToken)

	tokens2, err := cfg.GetOrCreateDoublePuppetTokens()
	require.NoError(t, err)
	assert.Equal(t, tokens, tokens2)
}

func TestConfig_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))

	cfg, err := NewDefault()
	require.NoError(t, err)
	cfg.EnableBridge("signal")

	err = cfg.Save()
	require.NoError(t, err)

	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.ServerName, loaded.ServerName)
	assert.Equal(t, cfg.Admin.Username, loaded.Admin.Username)
	assert.True(t, loaded.IsBridgeEnabled("signal"))
}

func TestConfig_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, ".local", "share"))

	assert.False(t, Exists())

	cfg, err := NewDefault()
	require.NoError(t, err)
	require.NoError(t, cfg.Save())

	assert.True(t, Exists())
}

func TestConfig_PublicBaseURL(t *testing.T) {
	cfg := &Config{ServerName: "localhost", ListenPort: 8008}
	assert.Equal(t, "http://localhost:8008", cfg.PublicBaseURL())
}

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(32)
	require.NoError(t, err)
	assert.Len(t, token, 32)

	// Two tokens should be different
	token2, err := GenerateToken(32)
	require.NoError(t, err)
	assert.NotEqual(t, token, token2)
}

func TestPaths(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))

	assert.Contains(t, ConfigDir(), filepath.Join(tmpDir, "config", "beetrix"))
	assert.Contains(t, DataDir(), filepath.Join(tmpDir, "data", "beetrix"))
	assert.Contains(t, SettingsPath(), "settings.yaml")
	assert.Contains(t, DendriteDir(), "dendrite")
	assert.Contains(t, BridgeDataDir("whatsapp"), filepath.Join("bridges", "whatsapp"))
	assert.Contains(t, SocketPath(), "beetrix.sock")
}

func TestEnsureDirs(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))

	require.NoError(t, EnsureDirs())

	_, err := os.Stat(ConfigDir())
	assert.NoError(t, err)
	_, err = os.Stat(DendriteDir())
	assert.NoError(t, err)
}
