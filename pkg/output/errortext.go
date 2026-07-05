package output

// SanitizeErrorText makes an error message safe to print to the terminal.
// API and OAuth error messages embed response-body text (Zuora reason
// messages, or raw non-JSON bodies such as a gateway HTML page), so the
// stderr error path needs the same terminal-injection defense the stdout
// table/detail path already has: newlines are preserved (errors are
// multi-line), tabs/CRs and the Unicode line/paragraph separators collapse
// to spaces, and other control characters (ANSI escapes) plus Unicode
// format characters (BiDi / zero-width, category Cf) are dropped. Invalid
// UTF-8 bytes (e.g. a multi-byte rune split by response-body truncation)
// render as U+FFFD instead of raw broken bytes.
func SanitizeErrorText(s string) string {
	return sanitizeRunes(s, true)
}

// SanitizeInline is SanitizeErrorText for single-line stderr contexts —
// progress/diagnostic lines that embed a response-derived value (e.g. a poll
// loop printing a job status). Newlines also collapse to spaces, matching
// the table-cell rules, so a hostile value cannot fake additional lines.
func SanitizeInline(s string) string {
	return sanitizeRunes(s, false)
}
