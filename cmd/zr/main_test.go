package main

import (
	"context"
	"errors"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/api"
)

func TestErrorLine_SanitizesTerminalEscapes(t *testing.T) {
	// An API error message embeds response-body text; a hostile body must not
	// be able to write raw escape sequences to the terminal via stderr.
	err := errors.New("Zuora API error (HTTP 502): \x1b[2Jspoofed\r\n  Message: a\u202eb")
	got := errorLine(err)
	want := "Error: Zuora API error (HTTP 502): [2Jspoofed \n  Message: ab"
	if got != want {
		t.Errorf("errorLine() = %q, want %q", got, want)
	}
}

// TestExitCode_APIErrorWrappingContextCanceledIs130 pins the exit-code tie:
// an APIError wrapping context.Canceled satisfies BOTH branches of exitCode
// (errors.Is via Unwrap, and exitCoder via ExitCode()==1 for StatusCode 0).
// The Ctrl-C check must win, so a cancelled request exits 130, not 1.
func TestExitCode_APIErrorWrappingContextCanceledIs130(t *testing.T) {
	err := &api.APIError{Err: context.Canceled}
	if got := exitCode(err); got != 130 {
		t.Errorf("exitCode(&api.APIError{Err: context.Canceled}) = %d, want 130", got)
	}
}

func TestErrorLine_PreservesMultiLineErrors(t *testing.T) {
	err := errors.New("Zuora API error (HTTP 400)\n  Code: 53100020\n  Message: bad value")
	got := errorLine(err)
	want := "Error: Zuora API error (HTTP 400)\n  Code: 53100020\n  Message: bad value"
	if got != want {
		t.Errorf("errorLine() = %q, want %q", got, want)
	}
}
