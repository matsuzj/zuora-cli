package cmdutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSleepContext_CompletesNormally(t *testing.T) {
	err := SleepContext(context.Background(), 10*time.Millisecond)
	assert.NoError(t, err)
}

func TestSleepContext_CancelledPromptly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := SleepContext(ctx, 5*time.Second)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	// A raw time.Sleep implementation would block the full 5s here.
	assert.Less(t, elapsed, 500*time.Millisecond, "cancellation must interrupt the sleep promptly")
}

func TestSleepContext_ZeroDuration(t *testing.T) {
	assert.NoError(t, SleepContext(context.Background(), 0))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.ErrorIs(t, SleepContext(ctx, 0), context.Canceled)
}
