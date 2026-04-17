package config

import (
	"os"
	"path/filepath"
)

const appName = "beetrix"

// ConfigDir returns the configuration directory (~/.config/beetrix/).
func ConfigDir() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, appName)
}

// DataDir returns the data directory (~/.local/share/beetrix/).
func DataDir() string {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, appName)
}

// SettingsPath returns the path to settings.yaml.
func SettingsPath() string {
	return filepath.Join(ConfigDir(), "settings.yaml")
}

// DendriteDir returns the Dendrite data directory.
func DendriteDir() string {
	return filepath.Join(DataDir(), "dendrite")
}

// BridgeDataDir returns the data directory for a specific bridge.
func BridgeDataDir(name string) string {
	return filepath.Join(DataDir(), "bridges", name)
}

// SocketPath returns the path to the Unix socket for IPC.
func SocketPath() string {
	return filepath.Join(DataDir(), "beetrix.sock")
}

// EnsureDirs creates all required directories.
func EnsureDirs() error {
	dirs := []string{
		ConfigDir(),
		DendriteDir(),
		filepath.Join(DataDir(), "bridges"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	return nil
}
