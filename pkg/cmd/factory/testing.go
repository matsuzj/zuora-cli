package factory

import (
	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
)

// NewTestFactory creates a Factory for testing with mock dependencies.
func NewTestFactory(ios *iostreams.IOStreams, cfg config.Config, baseURL, token string) *Factory {
	return &Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
		AuthToken: func() (string, error) {
			return token, nil
		},
		HttpClient: func() (*api.Client, error) {
			tokenFn := func() (string, error) { return token, nil }
			return api.NewClient(
				api.WithBaseURL(baseURL),
				api.WithTokenSource(tokenFn),
				api.WithRefreshToken(tokenFn),
				api.WithZuoraVersion(cfg.ZuoraVersion()),
			), nil
		},
	}
}

// NewTestFactoryReadOnly creates a Factory for testing with read-only mode enabled.
func NewTestFactoryReadOnly(ios *iostreams.IOStreams, cfg config.Config, baseURL, token string) *Factory {
	return &Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
		AuthToken: func() (string, error) {
			return token, nil
		},
		HttpClient: func() (*api.Client, error) {
			tokenFn := func() (string, error) { return token, nil }
			return api.NewClient(
				api.WithBaseURL(baseURL),
				api.WithTokenSource(tokenFn),
				api.WithRefreshToken(tokenFn),
				api.WithZuoraVersion(cfg.ZuoraVersion()),
				api.WithReadOnly(),
			), nil
		},
	}
}
