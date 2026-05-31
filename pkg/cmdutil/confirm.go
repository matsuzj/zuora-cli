package cmdutil

import "fmt"

// RequireConfirm enforces that the caller passed --confirm before an
// irreversible operation. It returns the canonical guard error when not
// confirmed, so every destructive command behaves consistently.
func RequireConfirm(confirmed bool) error {
	if !confirmed {
		return fmt.Errorf("this action is irreversible. Use --confirm to proceed")
	}
	return nil
}
