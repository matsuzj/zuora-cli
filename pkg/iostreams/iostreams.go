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
	if f, ok := s.Out.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// IsStderrTerminal returns true if stderr is a terminal (not piped).
func (s *IOStreams) IsStderrTerminal() bool {
	if f, ok := s.ErrOut.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// ColorEnabled reports whether colorized output should be used: only when
// stdout is a terminal and NO_COLOR is not set (https://no-color.org).
func (s *IOStreams) ColorEnabled() bool {
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor {
		return false
	}
	return s.IsTerminal()
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
