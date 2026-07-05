package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRoot is used by set_test.go and by tests that need direct cfg access.
func newTestRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	root.AddCommand(NewCmdConfig(f))
	return root
}

// newCmd adapts NewCmdConfig for use with cmdtest.Run (parent="").
func newCmd(f *factory.Factory) *cobra.Command { return NewCmdConfig(f) }

func TestConfigSet(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "set", "zuora_version", "2026-01-01"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Equal(t, "2026-01-01", cfg.ZuoraVersion())
	assert.Contains(t, out.String(), "Set zuora_version to 2026-01-01")
	assert.Equal(t, 1, cfg.SaveCallCount)
}

func TestConfigSet_InvalidKey(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newCmd, nil, "config", "set", "unknown_key", "value")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}

func TestConfigGet(t *testing.T) {
	stdout, _, err := cmdtest.Run(t, "", newCmd, nil, "config", "get", "active_environment")

	require.NoError(t, err)
	assert.Equal(t, "sandbox\n", stdout)
}

// TestConfigGet_KeyBranches covers the remaining runGet switch arms beyond
// active_environment (already tested above): zuora_version and default_output
// print the config values, and an unknown key is the exact error.
func TestConfigGet_KeyBranches(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"zuora_version", "2025-08-12\n"},
		{"default_output", "table\n"},
	}
	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			stdout, _, err := cmdtest.Run(t, "", newCmd, nil, "config", "get", tc.key)
			require.NoError(t, err)
			assert.Equal(t, tc.want, stdout)
		})
	}

	t.Run("unknown key", func(t *testing.T) {
		_, _, err := cmdtest.Run(t, "", newCmd, nil, "config", "get", "no_such_key")
		require.Error(t, err)
		assert.EqualError(t, err, "unknown config key: no_such_key")
	})
}

// completeConfig drives cobra's shell-completion machinery for
// `config <sub> <TAB>` and returns the raw completion output.
func completeConfig(t *testing.T, args ...string) string {
	t.Helper()
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")
	root := newTestRoot(f)
	root.SetOut(ios.Out)
	root.SetErr(ios.ErrOut)
	root.SetArgs(append([]string{"__complete", "config"}, args...))
	require.NoError(t, root.Execute(), "__complete must never fail")
	return out.String()
}

// TestConfigGetSet_CompleteKeys pins the ValidArgsFunction closures in
// newCmdGet and newCmdSet: the known config keys are suggested for the first
// argument with ShellCompDirectiveNoFileComp (:4), and nothing is suggested
// once a key is already present.
func TestConfigGetSet_CompleteKeys(t *testing.T) {
	for _, sub := range []string{"get", "set"} {
		t.Run(sub, func(t *testing.T) {
			lines := strings.Split(strings.TrimSpace(completeConfig(t, sub, "")), "\n")
			assert.Contains(t, lines, "active_environment")
			assert.Contains(t, lines, "zuora_version")
			assert.Contains(t, lines, "default_output")
			assert.Equal(t, ":4", lines[len(lines)-1], "directive must be ShellCompDirectiveNoFileComp")

			// Key already given: no further suggestions (set's <value> arg
			// and get's excess arg both hit the len(args)>0 branch).
			got := strings.TrimSpace(completeConfig(t, sub, "zuora_version", ""))
			assert.Equal(t, ":4", got, "no suggestions after the key argument")
		})
	}
}

func TestConfigList(t *testing.T) {
	stdout, _, err := cmdtest.Run(t, "", newCmd, nil, "config", "list")

	require.NoError(t, err)
	assert.Contains(t, stdout, "active_environment: sandbox")
	assert.Contains(t, stdout, "zuora_version: 2025-08-12")
	assert.Contains(t, stdout, "sandbox")
	assert.Contains(t, stdout, "rest.apisandbox.zuora.com")
}

// TestConfigList_JSON pins the output-consistency fix (#453): `config list
// --json` must emit structured JSON with a nested environments object, not the
// plain-text layout. cmdtest.Run registers and applies the global flags.
func TestConfigList_JSON(t *testing.T) {
	stdout, _, err := cmdtest.Run(t, "", newCmd, nil, "config", "list", "--json")
	require.NoError(t, err)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &got),
		"config list --json must emit valid JSON, not plain text")
	assert.Equal(t, "sandbox", got["active_environment"])
	assert.Equal(t, "2025-08-12", got["zuora_version"])
	envs, ok := got["environments"].(map[string]interface{})
	require.True(t, ok, "environments must be a nested object")
	sandbox, ok := envs["sandbox"].(map[string]interface{})
	require.True(t, ok, "sandbox environment must be present")
	assert.Equal(t, true, sandbox["active"])
	assert.Contains(t, sandbox["baseUrl"], "rest.apisandbox.zuora.com")
}

func TestConfigEnv(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "env", "us-production"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Equal(t, "us-production", cfg.ActiveEnvironment())
	assert.Contains(t, out.String(), "Switched to environment us-production")
}

func TestConfigEnv_Invalid(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newCmd, nil, "config", "env", "nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown environment: nonexistent")
}
