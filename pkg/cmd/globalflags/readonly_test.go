package globalflags

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
			assert.Equal(t, tc.want, EnvReadOnly(), "ZR_READ_ONLY=%q", tc.val)
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
	assert.False(t, EnvReadOnly())
}

func TestVerboseLevels(t *testing.T) {
	cases := []struct {
		count         int
		zrDebug       string
		verbose, body bool
	}{
		{0, "", false, false},
		{1, "", true, false},
		{2, "", true, true},
		{3, "", true, true},
		{0, "api", true, true}, // ZR_DEBUG=api implies both levels
		{0, "other", false, false},
	}
	for _, c := range cases {
		v, b := VerboseLevels(c.count, c.zrDebug)
		assert.Equal(t, c.verbose, v, "count=%d debug=%q", c.count, c.zrDebug)
		assert.Equal(t, c.body, b, "count=%d debug=%q", c.count, c.zrDebug)
	}
}
