package config

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the configuration directory path.
// Respects $XDG_CONFIG_HOME; falls back to ~/.config/zr/.
func configDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "zr")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "zr")
	}
	return filepath.Join(home, ".config", "zr")
}

func configFilePath(dir string) string {
	return filepath.Join(dir, "config.yml")
}

func environmentsFilePath(dir string) string {
	return filepath.Join(dir, "environments.yml")
}

func tokensFilePath(dir string) string {
	return filepath.Join(dir, "tokens.yml")
}
