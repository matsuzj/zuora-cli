package main

import (
	"errors"
	"testing"
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

func TestErrorLine_PreservesMultiLineErrors(t *testing.T) {
	err := errors.New("Zuora API error (HTTP 400)\n  Code: 53100020\n  Message: bad value")
	got := errorLine(err)
	want := "Error: Zuora API error (HTTP 400)\n  Code: 53100020\n  Message: bad value"
	if got != want {
		t.Errorf("errorLine() = %q, want %q", got, want)
	}
}
