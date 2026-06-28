// Package iostreams provides an abstraction over stdin/stdout/stderr
// with terminal detection.
package iostreams

import (
	"bytes"
	"io"
	"os"

	"golang.org/x/term"
)

// IOStreams provides access to standard I/O streams.
type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer

	// ttyOverride, when non-nil, forces IsTerminal()'s result. Set ONLY in tests
	// (via SetTTYForTest) to exercise the human/TTY branches — pager start and
	// the default_output-for-humans skip — that buffer-backed Test() streams
	// otherwise pin to non-TTY.
	ttyOverride *bool
}

// System returns IOStreams connected to the real terminal.
func System() *IOStreams {
	return &IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}

// IsTerminal returns true if stdout is a terminal (not piped).
func (s *IOStreams) IsTerminal() bool {
	if s.ttyOverride != nil {
		return *s.ttyOverride
	}
	if f, ok := s.Out.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// SetTTYForTest overrides IsTerminal() so tests can drive the human/TTY code
// paths (pager launch, default_output-for-humans) that are otherwise unreachable
// with buffer-backed streams. Test-only.
func (s *IOStreams) SetTTYForTest(isTTY bool) {
	s.ttyOverride = &isTTY
}

// Test returns IOStreams backed by bytes.Buffer for testing.
func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return &IOStreams{
		In:     io.NopCloser(in),
		Out:    out,
		ErrOut: errOut,
	}, in, out, errOut
}
