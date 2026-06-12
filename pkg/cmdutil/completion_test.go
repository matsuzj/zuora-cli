package cmdutil

import (
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnumCompletion_ReturnsValuesNoFileComp(t *testing.T) {
	fn := EnumCompletion("Alpha", "Beta")
	vals, directive := fn(nil, nil, "")
	assert.Equal(t, []string{"Alpha", "Beta"}, vals)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestEnvNamesCompletion_SortedNames(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.Envs["zeta"] = &config.Environment{}
	cfg.Envs["alpha"] = &config.Environment{}
	f := factory.NewTestFactory(ios, cfg, "http://example.invalid", "tok")

	vals, directive := EnvNamesCompletion(f)(nil, nil, "")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	require.GreaterOrEqual(t, len(vals), 2)
	assert.Contains(t, vals, "alpha")
	assert.Contains(t, vals, "zeta")
	assert.True(t, sortedStrings(vals), "names must be sorted")
}

func sortedStrings(v []string) bool {
	for i := 1; i < len(v); i++ {
		if v[i-1] > v[i] {
			return false
		}
	}
	return true
}

func TestEnvNamesCompletion_ConfigErrorDegrades(t *testing.T) {
	f := &factory.Factory{Config: func() (config.Config, error) {
		return nil, assert.AnError
	}}
	vals, directive := EnvNamesCompletion(f)(nil, nil, "")
	assert.Nil(t, vals)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}
