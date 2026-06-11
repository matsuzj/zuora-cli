package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetString(t *testing.T) {
	m := map[string]interface{}{
		"s":     "hello",
		"n":     float64(42),
		"big":   float64(1000000),
		"b":     true,
		"zero":  float64(0),
		"nilly": nil,
	}
	cases := map[string]string{
		"s":       "hello",
		"n":       "42",
		"big":     "1e+06", // GetString keeps fmt's default (scientific) notation
		"b":       "true",
		"zero":    "0",
		"nilly":   "",
		"missing": "",
	}
	for key, want := range cases {
		if got := GetString(m, key); got != want {
			t.Errorf("GetString(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestGetDecimal(t *testing.T) {
	m := map[string]interface{}{
		"s":     "hello",
		"amt":   float64(1000000),
		"cents": float64(12.5),
		"zero":  float64(0),
		"nilly": nil,
	}
	cases := map[string]string{
		"s":       "hello",
		"amt":     "1000000", // plain decimal, not 1e+06
		"cents":   "12.5",
		"zero":    "0",
		"nilly":   "",
		"missing": "",
	}
	for key, want := range cases {
		if got := GetDecimal(m, key); got != want {
			t.Errorf("GetDecimal(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestGetMoney(t *testing.T) {
	m := map[string]interface{}{"bal": 50.0, "neg": -10.5, "str": "n/a"}
	assert.Equal(t, "50.00", GetMoney(m, "bal"), "two decimals preserved")
	assert.Equal(t, "-10.50", GetMoney(m, "neg"))
	assert.Equal(t, "n/a", GetMoney(m, "str"), "non-float falls back to %v")
	assert.Equal(t, "", GetMoney(m, "missing"))
	assert.Equal(t, "", GetMoney(map[string]interface{}{"bal": nil}, "bal"))
}

func TestGetBool(t *testing.T) {
	m := map[string]interface{}{"on": true, "off": false, "str": "true"}
	assert.Equal(t, "true", GetBool(m, "on"))
	assert.Equal(t, "false", GetBool(m, "off"))
	assert.Equal(t, "", GetBool(m, "str"), "non-bool is empty, not coerced")
	assert.Equal(t, "", GetBool(m, "missing"))
}

func TestGetInt(t *testing.T) {
	m := map[string]interface{}{"day": 15.0, "frac": 15.9, "str": "15"}
	assert.Equal(t, "15", GetInt(m, "day"))
	assert.Equal(t, "15", GetInt(m, "frac"), "truncated, not rounded")
	assert.Equal(t, "", GetInt(m, "str"))
	assert.Equal(t, "", GetInt(m, "missing"))
}

func TestAddBodyAndConfirmFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "x"}
	var body string
	var confirm bool
	AddBodyFlag(cmd, &body, true)
	AddConfirmFlag(cmd, &confirm, "deletion")

	bf := cmd.Flags().Lookup("body")
	require.NotNil(t, bf)
	assert.Equal(t, "b", bf.Shorthand)
	assert.Equal(t, "Request body (JSON string, @file, or - for stdin)", bf.Usage)

	cf := cmd.Flags().Lookup("confirm")
	require.NotNil(t, cf)
	assert.Equal(t, "Confirm the deletion", cf.Usage)
}
