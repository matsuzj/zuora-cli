package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// newCancelledContext returns a context that is already cancelled, plus its
// cancel func (so callers can defer it without a lint complaint).
func newCancelledContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx, cancel
}

// AuthError.Error must render the hint on its own indented line when a hint is
// present, and must render the bare message (no trailing newline/indent) when it
// is absent. These are the two branches of Error(); the no-hint branch is what
// callers rely on so a hintless error does not print a stray blank guidance line.
func TestAuthError_Error_WithAndWithoutHint(t *testing.T) {
	withHint := &AuthError{Message: "authentication failed", Hint: "Check your Client ID."}
	assert.Equal(t, "authentication failed\n  Check your Client ID.", withHint.Error(),
		"a hint must appear on a second, indented line")

	withoutHint := &AuthError{Message: "authentication failed"}
	assert.Equal(t, "authentication failed", withoutHint.Error(),
		"with no hint the message must be returned verbatim, with no newline or indent")
}
