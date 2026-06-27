package output

import (
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
