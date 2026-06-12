package globalflags_test

import (
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/globalflags"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// applyWithZREnv runs Register+Apply with ZR_ENV set, returning the factory.
func applyWithZREnv(t *testing.T, zrEnv string, args ...string) *factory.Factory {
	t.Helper()
	t.Setenv("ZR_ENV", zrEnv)

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.Envs["alpha"] = &config.Environment{BaseURL: "https://rest.test.zuora.com"}
	cfg.Envs["beta"] = &config.Environment{BaseURL: "https://rest.test.zuora.com"}
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "tok")

	cmd := &cobra.Command{Use: "x"}
	globalflags.Register(cmd)
	require.NoError(t, cmd.ParseFlags(args))
	require.NoError(t, globalflags.Apply(f, cmd))
	return f
}

// ZR_ENV overrides the config's active environment (P6-4).
func TestApply_ZREnvOverridesActiveEnvironment(t *testing.T) {
	f := applyWithZREnv(t, "beta")
	cfg, err := f.Config()
	require.NoError(t, err)
	assert.Equal(t, "beta", cfg.ActiveEnvironment())
}

// The --env flag wins over ZR_ENV (same precedence shape as --read-only).
func TestApply_EnvFlagWinsOverZREnv(t *testing.T) {
	f := applyWithZREnv(t, "beta", "--env", "alpha")
	cfg, err := f.Config()
	require.NoError(t, err)
	assert.Equal(t, "alpha", cfg.ActiveEnvironment())
}

// An unknown ZR_ENV errors exactly like an unknown --env — a typo must not
// silently fall back to the config default (it could target another tenant).
func TestApply_UnknownZREnvErrors(t *testing.T) {
	f := applyWithZREnv(t, "no-such-env")
	_, err := f.Config()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid environment "no-such-env"`)
}
