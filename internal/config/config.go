// Package config manages CLI configuration files.
package config

import "time"

// Config provides access to CLI configuration.
type Config interface {
	// ActiveEnvironment returns the current environment name.
	ActiveEnvironment() string
	// SetActiveEnvironment changes the active environment.
	SetActiveEnvironment(name string) error
	// ZuoraVersion returns the default Zuora API version header.
	ZuoraVersion() string
	// SetZuoraVersion changes the default API version.
	SetZuoraVersion(v string) error
	// DefaultOutput returns the default output format (table, json).
	DefaultOutput() string
	// SetDefaultOutput changes the default output format.
	SetDefaultOutput(v string) error

	// Environment returns the environment configuration by name.
	Environment(name string) (*Environment, error)
	// Environments returns all configured environments.
	Environments() map[string]*Environment
	// AddEnvironment adds or updates an environment.
	AddEnvironment(name string, env *Environment) error
	// RemoveEnvironment removes an environment.
	RemoveEnvironment(name string) error

	// Token returns the cached token for an environment.
	Token(envName string) (*TokenEntry, error)
	// SetToken stores a token for an environment.
	SetToken(envName string, token *TokenEntry) error
	// RemoveToken removes the cached token for an environment.
	RemoveToken(envName string) error

	// ConfigDir returns the configuration directory path.
	ConfigDir() string
	// Save persists all configuration changes to disk.
	Save() error
}

// Environment represents a Zuora environment endpoint.
type Environment struct {
	BaseURL string `yaml:"base_url"`
}

// TokenEntry represents a cached OAuth token.
type TokenEntry struct {
	AccessToken string    `yaml:"access_token"`
	ExpiresAt   time.Time `yaml:"expires_at"`
}

// IsValid returns true if the token is not expired (with 60s buffer).
func (t *TokenEntry) IsValid() bool {
	if t == nil || t.AccessToken == "" {
		return false
	}
	return time.Now().Add(60 * time.Second).Before(t.ExpiresAt)
}
