package output

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
)

// nopWriteCloser wraps an io.Writer as an io.WriteCloser with a no-op Close.
type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }

// StartPager starts a pager process if stdout is a TTY.
// Returns an io.WriteCloser to write to. Call Close() when done.
// If not a TTY or pager is unavailable, returns a no-op wrapper around ios.Out.
func StartPager(ios *iostreams.IOStreams) (io.WriteCloser, error) {
	if !ios.IsTerminal() {
		return nopWriteCloser{ios.Out}, nil
	}

	pagerCmd := os.Getenv("PAGER")
	if pagerCmd == "" {
		pagerCmd = "less"
	}
	if pagerCmd == "cat" {
		return nopWriteCloser{ios.Out}, nil
	}

	// Split PAGER into program + args to support values like "less -FRX"
	parts := strings.Fields(pagerCmd)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = ios.Out
	cmd.Stderr = ios.ErrOut

	// Set LESS/LV defaults like gh CLI so less auto-exits when content fits
	cmd.Env = os.Environ()
	if os.Getenv("LESS") == "" {
		cmd.Env = append(cmd.Env, "LESS=FRX")
	}
	if os.Getenv("LV") == "" {
		cmd.Env = append(cmd.Env, "LV=-c")
	}

	w, err := cmd.StdinPipe()
	if err != nil {
		return nopWriteCloser{ios.Out}, err
	}

	if err := cmd.Start(); err != nil {
		return nopWriteCloser{ios.Out}, err
	}

	return &pagerWriteCloser{pipe: w, cmd: cmd}, nil
}

type pagerWriteCloser struct {
	pipe io.WriteCloser
	cmd  *exec.Cmd
}

func (p *pagerWriteCloser) Write(b []byte) (int, error) {
	return p.pipe.Write(b)
}

func (p *pagerWriteCloser) Close() error {
	p.pipe.Close()
	return p.cmd.Wait()
}
