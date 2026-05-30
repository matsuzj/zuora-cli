package root

import (
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
	// With the var unset entirely, read-only must be off.
	assert.False(t, envReadOnly())
}
