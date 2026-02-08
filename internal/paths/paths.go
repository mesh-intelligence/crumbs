// Package paths provides platform-specific directory path handling.
// Implements: prd010-configuration-directories R1, R2;
//
//	docs/ARCHITECTURE § Configuration.
package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

// DefaultConfigDir returns the platform-specific default configuration directory.
// Per prd010-configuration-directories R1.2:
//   - Linux: $XDG_CONFIG_HOME/crumbs (falls back to ~/.config/crumbs)
//   - macOS: ~/Library/Application Support/crumbs
//   - Windows: %APPDATA%\crumbs
func DefaultConfigDir() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "crumbs"), nil

	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "crumbs"), nil

	default: // Linux and other Unix-like systems
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			xdgConfig = filepath.Join(home, ".config")
		}
		return filepath.Join(xdgConfig, "crumbs"), nil
	}
}

// DefaultDataDir returns the platform-specific default data directory.
// Per prd010-configuration-directories R2.2:
//   - Linux: $XDG_DATA_HOME/crumbs (falls back to ~/.local/share/crumbs)
//   - macOS: ~/Library/Application Support/crumbs/data
//   - Windows: %LOCALAPPDATA%\crumbs
func DefaultDataDir() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "crumbs", "data"), nil

	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "crumbs"), nil

	default: // Linux and other Unix-like systems
		xdgData := os.Getenv("XDG_DATA_HOME")
		if xdgData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			xdgData = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(xdgData, "crumbs"), nil
	}
}

// ResolveConfigDir resolves the configuration directory with precedence:
// flag > env > platform default (per R1.3).
func ResolveConfigDir(flagValue, envVar string) (string, error) {
	// Highest precedence: CLI flag
	if flagValue != "" {
		return flagValue, nil
	}

	// Middle precedence: environment variable
	if envVar != "" {
		envValue := os.Getenv(envVar)
		if envValue != "" {
			return envValue, nil
		}
	}

	// Lowest precedence: platform default
	return DefaultConfigDir()
}

// ResolveDataDir resolves the data directory with precedence:
// flag > config file value > platform default (per R2.3).
func ResolveDataDir(flagValue, configValue string) (string, error) {
	// Highest precedence: CLI flag
	if flagValue != "" {
		return flagValue, nil
	}

	// Middle precedence: config file value
	if configValue != "" {
		return configValue, nil
	}

	// Lowest precedence: platform default
	return DefaultDataDir()
}

// EnsureDir creates a directory if it does not exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
