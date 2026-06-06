package cmdutil

import "testing"

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
