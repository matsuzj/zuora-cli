package output

import (
	"io"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A jq program that raises a runtime error (error("...")) must surface a
// non-nil "jq error" so a failing filter produces a non-zero exit code rather
// than silently emitting an empty stream.
func TestPrintJSON_JQ_RuntimeError(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	err := PrintJSON(ios, []byte(`{}`), `.x|error("boom")`)
	require.Error(t, err, "a jq runtime error must be returned")
	assert.Contains(t, err.Error(), "jq error", "error must be tagged as a jq error")
	assert.Contains(t, err.Error(), "boom", "the underlying jq message must be preserved")
	assert.Empty(t, out.String(), "nothing should be written to stdout when jq errors")
}

// halt_error carrying a value must surface a "jq halt" error (a *gojq.HaltError
// whose Value() is non-nil), distinct from an ordinary runtime error, so scripts
// can tell a deliberate halt from a filter failure. Piping a value into
// halt_error is the form that produces a value-bearing HaltError in gojq.
func TestPrintJSON_JQ_HaltErrorWithValue(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	err := PrintJSON(ios, []byte(`{}`), `5|halt_error`)
	require.Error(t, err, "halt_error with a value must be returned")
	assert.Contains(t, err.Error(), "jq halt", "error must be tagged as a jq halt, not a generic jq error")
	assert.Contains(t, err.Error(), "5", "the halt value must be preserved")
	assert.Empty(t, out.String())
}

// A valueless `halt` must stop iteration WITHOUT producing an error and without
// writing partial output, so a clean early termination is not reported as a
// failure.
func TestPrintJSON_JQ_PlainHalt_NoError(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	err := PrintJSON(ios, []byte(`{"a":1}`), `halt`)
	require.NoError(t, err, "a valueless halt is a clean stop, not an error")
	assert.Empty(t, out.String(), "halt before emitting any value must print nothing")
}

// On a non-terminal stream StartPager must NOT spawn a pager: it returns a
// working no-op WriteCloser whose writes reach ios.Out and whose Close is a
// nil-returning no-op. This is the branch the table renderer relies on when
// stdout is piped/redirected.
func TestStartPager_NonTerminal_NoOpWriter(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	require.False(t, ios.IsTerminal(), "test streams must be non-terminal for this branch")

	w, err := StartPager(ios)
	require.NoError(t, err)
	require.NotNil(t, w)

	n, err := w.Write([]byte("hello pager"))
	require.NoError(t, err)
	assert.Equal(t, len("hello pager"), n)
	assert.Equal(t, "hello pager", out.String(), "writes must pass through to ios.Out unpaged")

	assert.NoError(t, w.Close(), "Close on the no-op wrapper must be a nil-returning no-op")
}

// The returned no-op WriteCloser must genuinely wrap ios.Out (not a discarded
// buffer): a second write after Close still appends to the same stream, proving
// no real pager process owns the pipe.
func TestStartPager_NonTerminal_WrapsOut(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	w, err := StartPager(ios)
	require.NoError(t, err)
	var wc io.WriteCloser = w
	_, _ = wc.Write([]byte("a"))
	_ = wc.Close()
	_, err = wc.Write([]byte("b"))
	require.NoError(t, err)
	assert.Equal(t, "ab", out.String())
}

// TestPrintJSON_NonJSON_StderrSanitized pins that the invalid-JSON fallback
// (e.g. a gateway HTML page) is sanitized before being echoed to stderr — the
// same terminal-injection defense as the table/detail/error paths.
func TestPrintJSON_NonJSON_StderrSanitized(t *testing.T) {
	ios, _, out, errOut := iostreams.Test()
	err := PrintJSON(ios, []byte("<html>\x1b[2Jspoofed\x1b[H</html>"), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not valid JSON")
	assert.Empty(t, out.String(), "invalid JSON must not reach stdout")
	assert.NotContains(t, errOut.String(), "\x1b", "stderr echo must not carry escape codes")
	assert.Contains(t, errOut.String(), "[2Jspoofed[H", "body text must still be visible for diagnosis")
}
