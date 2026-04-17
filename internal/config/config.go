package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the beetrix settings.
type Config struct {
	ServerName     string                  `yaml:"server_name"`
	ListenAddress  string                  `yaml:"listen_address"`
	ListenPort     int                     `yaml:"listen_port"`
	Federation     bool                    `yaml:"federation"`
	LogLevel       string                  `yaml:"log_level"`
	Admin          AdminConfig             `yaml:"admin"`
	EnabledBridges []string                `yaml:"enabled_bridges"`
	BridgeTokens   map[string]BridgeTokens `yaml:"bridge_tokens,omitempty"`
	Telegram       *TelegramConfig         `yaml:"telegram,omitempty"`
	DoublePuppet   *BridgeTokens           `yaml:"double_puppet_tokens,omitempty"`
	ExternalBridges map[string]ExternalBridgeConfig `yaml:"external_bridges,omitempty"`
}

// AdminConfig holds admin user credentials.
type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// TelegramConfig holds Telegram API credentials.
type TelegramConfig struct {
	APIID   string `yaml:"api_id"`
	APIHash string `yaml:"api_hash"`
}

// BridgeTokens holds the appservice tokens for a bridge.
type BridgeTokens struct {
	ASToken string `yaml:"as_token"`
	HSToken string `yaml:"hs_token"`
}

// ExternalBridgeConfig holds configuration for an external bridge subprocess.
type ExternalBridgeConfig struct {
	Binary string `yaml:"binary"`
	Port   int    `yaml:"port"`
}

// Port returns the listen port with a default of 8008.
func (c *Config) Port() int {
	if c.ListenPort != 0 {
		return c.ListenPort
	}
	return 8008
}

// Address returns the listen address with a default of 127.0.0.1.
func (c *Config) Address() string {
	if c.ListenAddress != "" {
		return c.ListenAddress
	}
	return "127.0.0.1"
}

// IsBridgeEnabled checks if a specific bridge is enabled.
func (c *Config) IsBridgeEnabled(name string) bool {
	for _, b := range c.EnabledBridges {
		if b == name {
			return true
		}
	}
	return false
}

// EnableBridge adds a bridge to the enabled list.
func (c *Config) EnableBridge(name string) {
	if !c.IsBridgeEnabled(name) {
		c.EnabledBridges = append(c.EnabledBridges, name)
	}
}

// DisableBridge removes a bridge from the enabled list.
func (c *Config) DisableBridge(name string) {
	filtered := make([]string, 0, len(c.EnabledBridges))
	for _, b := range c.EnabledBridges {
		if b != name {
			filtered = append(filtered, b)
		}
	}
	c.EnabledBridges = filtered
}

// GetOrCreateBridgeTokens returns existing tokens or creates new ones.
func (c *Config) GetOrCreateBridgeTokens(bridgeName string) (BridgeTokens, error) {
	if c.BridgeTokens == nil {
		c.BridgeTokens = make(map[string]BridgeTokens)
	}
	if tokens, ok := c.BridgeTokens[bridgeName]; ok && tokens.ASToken != "" && tokens.HSToken != "" {
		return tokens, nil
	}

	asToken, err := GenerateToken(64)
	if err != nil {
		return BridgeTokens{}, err
	}
	hsToken, err := GenerateToken(64)
	if err != nil {
		return BridgeTokens{}, err
	}

	tokens := BridgeTokens{ASToken: asToken, HSToken: hsToken}
	c.BridgeTokens[bridgeName] = tokens
	return tokens, nil
}

// GetOrCreateDoublePuppetTokens returns existing tokens or creates new ones.
func (c *Config) GetOrCreateDoublePuppetTokens() (BridgeTokens, error) {
	if c.DoublePuppet != nil && c.DoublePuppet.ASToken != "" && c.DoublePuppet.HSToken != "" {
		return *c.DoublePuppet, nil
	}

	asToken, err := GenerateToken(64)
	if err != nil {
		return BridgeTokens{}, err
	}
	hsToken, err := GenerateToken(64)
	if err != nil {
		return BridgeTokens{}, err
	}

	tokens := BridgeTokens{ASToken: asToken, HSToken: hsToken}
	c.DoublePuppet = &tokens
	return tokens, nil
}

// PublicBaseURL returns the public URL for the homeserver.
func (c *Config) PublicBaseURL() string {
	return fmt.Sprintf("http://%s:%d", c.ServerName, c.Port())
}

// NewDefault creates a config with sensible defaults and generated secrets.
func NewDefault() (*Config, error) {
	adminPass, err := GenerateToken(16)
	if err != nil {
		return nil, err
	}

	return &Config{
		ServerName:     "localhost",
		ListenAddress:  "127.0.0.1",
		ListenPort:     8008,
		Federation:     false,
		LogLevel:       "info",
		Admin: AdminConfig{
			Username: "admin",
			Password: adminPass,
		},
		EnabledBridges: []string{},
	}, nil
}

// Load reads the config from settings.yaml.
func Load() (*Config, error) {
	data, err := os.ReadFile(SettingsPath())
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to settings.yaml.
func (c *Config) Save() error {
	if err := EnsureDirs(); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(SettingsPath(), data, 0600)
}

// Exists checks if a config file already exists.
func Exists() bool {
	_, err := os.Stat(SettingsPath())
	return err == nil
}

// GenerateToken creates a cryptographically secure random token.
func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// IsPortAvailable checks if a port is available for use.
func IsPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}
