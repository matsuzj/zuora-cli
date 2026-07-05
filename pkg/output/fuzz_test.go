package output

import (
	"testing"
	"unicode"
	"unicode/utf8"
)

// FuzzSanitizeRunes pins the terminal-injection defense shared by the
// table/detail/CSV/error-text paths: for ANY input (including invalid
// UTF-8), the output carries no escape or format characters capable of
// driving a terminal, is always valid UTF-8, and sanitizing twice changes
// nothing (idempotence — a second pass must not reveal new characters).
func FuzzSanitizeRunes(f *testing.F) {
	seeds := []string{
		"",
		"plain text",
		"a\x1b[31mred\x1b[0m",
		"x\x1b]0;title\x07y",
		"a\u202eb\u200bc",
		"line1\r\nline2\ttabbed",
		"a\u2028b\u2029c",
		"trunc\xe3",
		"\x9b31m", // C1 CSI single byte
		"多バイト\x1b[2J混在",
	}
	for _, s := range seeds {
		f.Add(s, true)
		f.Add(s, false)
	}
	f.Fuzz(func(t *testing.T, s string, preserveNewlines bool) {
		out := sanitizeRunes(s, preserveNewlines)
		if !utf8.ValidString(out) {
			t.Errorf("output is not valid UTF-8 for input %q", s)
		}
		for _, r := range out {
			if r == '\n' {
				if !preserveNewlines {
					t.Errorf("newline survived with preserveNewlines=false (input %q)", s)
				}
				continue
			}
			if r == '\t' || r == '\r' || r == '\u2028' || r == '\u2029' {
				t.Errorf("separator %U survived (input %q)", r, s)
			}
			if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
				t.Errorf("control/format rune %U survived (input %q)", r, s)
			}
		}
		if again := sanitizeRunes(out, preserveNewlines); again != out {
			t.Errorf("not idempotent for input %q: %q -> %q", s, out, again)
		}
	})
}
