package factory

import (
	"context"
	"time"

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
		AuthToken: func(context.Context) (string, error) {
			return token, nil
		},
		HttpClient: func() (*api.Client, error) {
			tokenFn := func(context.Context) (string, error) { return token, nil }
			return api.NewClient(
				api.WithBaseURL(baseURL),
				api.WithTokenSource(tokenFn),
				api.WithRefreshToken(tokenFn),
				api.WithZuoraVersion(cfg.ZuoraVersion()),
				// No real backoff in tests: a handler that returns a retryable
				// status (429/5xx) must not silently sleep 1-6s per retry.
				// Cancellation is still honored, matching sleepWithContext.
				api.WithSleep(func(ctx context.Context, _ time.Duration) error {
					return ctx.Err()
				}),
			), nil
		},
	}
}
