// Package factory provides dependency injection for CLI commands.
package factory

import (
	"sync"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/internal/auth"
	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
)

// Factory provides shared dependencies to all commands.
type Factory struct {
	IOStreams   *iostreams.IOStreams
	Config     func() (config.Config, error)
	HttpClient func() (*api.Client, error)
	AuthToken  func() (string, error)
}

// New creates a Factory with real (system) dependencies.
// Config, HttpClient, and AuthToken are lazily initialized.
func New() *Factory {
	f := &Factory{
		IOStreams: iostreams.System(),
	}

	// Lazy config — loaded once, cached
	var configOnce sync.Once
	var cachedConfig config.Config
	var configErr error
	f.Config = func() (config.Config, error) {
		configOnce.Do(func() {
			cachedConfig, configErr = config.LoadDefault()
		})
		return cachedConfig, configErr
	}

	// Lazy auth token
	f.AuthToken = func() (string, error) {
		cfg, err := f.Config()
		if err != nil {
			return "", err
		}
		creds := auth.NewCredentialStore()
		ts := &auth.TokenSource{Config: cfg, Creds: creds}
		return ts.Token(cfg.ActiveEnvironment())
	}

	// Lazy HTTP client
	f.HttpClient = func() (*api.Client, error) {
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}
		env, err := cfg.Environment(cfg.ActiveEnvironment())
		if err != nil {
			return nil, err
		}
		// refreshToken forces a token refresh (bypasses cache)
		refreshToken := func() (string, error) {
			creds := auth.NewCredentialStore()
			ts := &auth.TokenSource{Config: cfg, Creds: creds}
			return ts.Refresh(cfg.ActiveEnvironment())
		}
		return api.NewClient(
			api.WithBaseURL(env.BaseURL),
			api.WithTokenSource(f.AuthToken),
			api.WithRefreshToken(refreshToken),
			api.WithZuoraVersion(cfg.ZuoraVersion()),
		), nil
	}

	return f
}
