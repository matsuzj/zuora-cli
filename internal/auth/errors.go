// Package auth provides OAuth 2.0 authentication for Zuora APIs.
package auth

import "fmt"

// AuthError represents an authentication failure.
type AuthError struct {
	Message string
	Hint    string
	// StatusCode is the HTTP status from the OAuth server when the failure came
	// from an HTTP response (0 otherwise).
	StatusCode int
}

func (e *AuthError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s\n  %s", e.Message, e.Hint)
	}
	return e.Message
}

// ExitCode returns the exit code for authentication errors. A 5xx from the OAuth
// server is a server-side failure (exit 4, matching APIError); anything else is
// treated as a credential/auth error (exit 2).
func (e *AuthError) ExitCode() int {
	if e.StatusCode >= 500 {
		return 4
	}
	return 2
}
