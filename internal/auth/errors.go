// Package auth provides OAuth 2.0 authentication for Zuora APIs.
package auth

import "fmt"

// AuthError represents an authentication failure.
type AuthError struct {
	Message string
	Hint    string
}

func (e *AuthError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s\n  %s", e.Message, e.Hint)
	}
	return e.Message
}

// ExitCode returns the exit code for authentication errors.
func (e *AuthError) ExitCode() int { return 2 }
