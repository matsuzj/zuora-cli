package output

import "testing"

func TestSanitizeErrorText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"plain multi-line error preserved", "Zuora API error (HTTP 400)\n  Code: 53100020\n  Message: bad value", "Zuora API error (HTTP 400)\n  Code: 53100020\n  Message: bad value"},
		{"ANSI escape dropped", "bad \x1b[31mred\x1b[0m value", "bad [31mred[0m value"},
		{"OSC escape dropped", "x\x1b]0;spoofed title\x07y", "x]0;spoofed titley"},
		{"carriage return to space", "line1\rline2", "line1 line2"},
		{"tab to space", "a\tb", "a b"},
		{"unicode line separators to space", "a\u2028b\u2029c", "a b c"},
		{"BiDi override dropped", "abc\u202edef", "abcdef"},
		{"zero-width space dropped", "a\u200bb", "ab"},
		{"invalid utf8 becomes replacement rune", "trunc\xe3", "trunc�"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeErrorText(tt.in); got != tt.want {
				t.Errorf("SanitizeErrorText(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSanitizeInline(t *testing.T) {
	// Single-line variant: escapes/format chars drop AND newlines collapse,
	// so a hostile value cannot fake extra lines in a progress message.
	if got := SanitizeInline("In\x1b[31m Progress\nDone"); got != "In[31m Progress Done" {
		t.Errorf("SanitizeInline() = %q", got)
	}
}
