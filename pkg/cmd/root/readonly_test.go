package root

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvReadOnly(t *testing.T) {
	cases := []struct {
		val  string
		want bool
	}{
		{"", false},
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"1", true},
		{"yes", true},
		{"y", true},
		{"on", true},
		{"t", true},
		{"false", false},
		{"FALSE", false},
		{"0", false},
		{"no", false},
		{"off", false},
		{"n", false},
		// Fail-safe: an unrecognized non-empty value enables read-only.
		{"maybe", true},
		{"  true  ", true},
	}
	for _, tc := range cases {
		t.Run(tc.val, func(t *testing.T) {
			t.Setenv("ZR_READ_ONLY", tc.val)
			assert.Equal(t, tc.want, envReadOnly(), "ZR_READ_ONLY=%q", tc.val)
		})
	}
}

func TestEnvReadOnly_Unset(t *testing.T) {
	// With the var unset entirely, read-only must be off. t.Setenv overrides
	// any ambient ZR_READ_ONLY (so a developer machine exporting it — a
	// documented safety setup — cannot flip this assertion) and registers
	// restoration of the pre-test state; it also guards against t.Parallel.
	// envReadOnly currently reads os.Getenv, which cannot tell "" from unset,
	// so t.Setenv alone would pass today — the extra os.Unsetenv makes the
	// test true to its name and keeps it valid if envReadOnly ever switches
	// to os.LookupEnv.
	t.Setenv("ZR_READ_ONLY", "")
	os.Unsetenv("ZR_READ_ONLY")
	assert.False(t, envReadOnly())
}
