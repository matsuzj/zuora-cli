package cmdutil

import (
	"context"
	"time"
)

// SleepContext waits for d or until ctx is cancelled, whichever comes first,
// returning ctx.Err() on cancellation. Polling loops must use this instead of
// time.Sleep: a raw sleep holds Ctrl-C hostage for the full interval (the
// signal context is cancelled but nothing observes it until the sleep ends).
func SleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
