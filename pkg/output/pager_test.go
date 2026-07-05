package output

import (
	"bytes"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartPager_TTYBranchReachable(t *testing.T) {
	// Non-TTY: StartPager returns a passthrough writer (no pager) — writes reach Out.
	ios, _, out, _ := iostreams.Test()
	w, err := StartPager(ios)
	require.NoError(t, err)
	_, _ = w.Write([]byte("piped"))
	require.NoError(t, w.Close())
	assert.Equal(t, "piped", out.String(), "non-TTY writes straight to Out")

	// The IsTerminal()==true path is now reachable via the hook (F-25). PAGER=cat
	// takes StartPager's no-child-process fast path, so the TTY code runs without
	// launching an interactive pager.
	t.Setenv("PAGER", "cat")
	ios2, _, out2, _ := iostreams.Test()
	ios2.SetTTYForTest(true)
	w2, err := StartPager(ios2)
	require.NoError(t, err)
	_, _ = w2.Write([]byte("tty"))
	require.NoError(t, w2.Close())
	assert.Equal(t, "tty", out2.String(), "TTY + PAGER=cat path writes to Out without a child process")
}

// TestStartPager_EarlyQuitSwallowsEPIPE pins the user-quits-pager-early
// contract: once the pager process exits (like pressing 'q' in less),
// further Writes hit EPIPE, which pagerWriteCloser.Write must swallow, and
// Close must not surface the dead pipe either. `true` exits immediately
// without reading stdin, so pushing a few MiB through the pipe guarantees
// the EPIPE path actually runs.
func TestStartPager_EarlyQuitSwallowsEPIPE(t *testing.T) {
	t.Setenv("PAGER", "true") // exits at once, never reads stdin
	ios, _, _, _ := iostreams.Test()
	ios.SetTTYForTest(true)

	w, err := StartPager(ios)
	require.NoError(t, err, "spawning the pager must succeed")

	chunk := bytes.Repeat([]byte("x"), 64*1024)
	const chunks = 64 // 4 MiB total — far beyond any OS pipe buffer
	reported := 0
	for i := 0; i < chunks; i++ {
		n, werr := w.Write(chunk)
		require.NoError(t, werr, "EPIPE from a dead pager must be swallowed, not surfaced")
		reported += n
	}
	// Non-vacuousness guard: the pipe must actually have died mid-stream —
	// otherwise the EPIPE swallow in Write was never exercised.
	assert.Less(t, reported, chunks*len(chunk),
		"the dead pager's pipe should have rejected most of the data via EPIPE")

	require.NoError(t, w.Close(), "Close must ignore the early-exited pager")
}

// TestStartPager_SpawnFailureFallsBackToOut pins the unstartable-PAGER
// contract: StartPager returns the error AND a usable fallback writer that
// goes straight to ios.Out, so callers can report the problem yet still
// render output.
func TestStartPager_SpawnFailureFallsBackToOut(t *testing.T) {
	t.Setenv("PAGER", "/nonexistent/zzz-no-such-pager")
	ios, _, out, _ := iostreams.Test()
	ios.SetTTYForTest(true)

	w, err := StartPager(ios)
	require.Error(t, err, "an unstartable PAGER must surface the spawn error")
	require.NotNil(t, w, "…while still returning a usable fallback writer")

	_, werr := w.Write([]byte("fallback"))
	require.NoError(t, werr)
	require.NoError(t, w.Close())
	assert.Equal(t, "fallback", out.String(), "fallback writer must reach ios.Out directly")
}

// TestStartPager_MultiWordPagerCommand pins the strings.Fields split of
// PAGER values like "less -FRX": the first word is the program, the rest are
// its args. "cat -u" passes stdin through, so the written bytes must arrive
// on ios.Out (exec.Cmd copies the child's stdout into the buffer and Close's
// cmd.Wait joins that copy).
func TestStartPager_MultiWordPagerCommand(t *testing.T) {
	t.Setenv("PAGER", "cat -u")
	ios, _, out, _ := iostreams.Test()
	ios.SetTTYForTest(true)

	w, err := StartPager(ios)
	require.NoError(t, err, `PAGER="cat -u" must be split into program+args, not exec'd as one path`)

	_, werr := w.Write([]byte("through the pager\n"))
	require.NoError(t, werr)
	require.NoError(t, w.Close())
	assert.Equal(t, "through the pager\n", out.String(), "cat must have passed the bytes through to Out")
}
