package jobstatus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/globalflags"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdJobStatus(f) }

func TestOrderJobStatus_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/async-jobs/2c92c0f9876", map[string]interface{}{
		"success":       true,
		"jobId":         "2c92c0f9876",
		"status":        "Completed",
		"result":        "Success",
		"orderNumber":   "O-00000001",
		"accountNumber": "A001",
	})

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "job-status", "2c92c0f9876")
	require.NoError(t, err)
	assert.Contains(t, stdout, "2c92c0f9876")
	assert.Contains(t, stdout, "Completed")
	assert.Contains(t, stdout, "O-00000001")
}

func TestOrderJobStatus_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "job-status")
	assert.Error(t, err)
}

// Always-InProgress server for watch tests.
func inProgressHandler(t *testing.T, calls *int32) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(calls, 1)
		w.WriteHeader(200)
		status := "InProgress"
		if n >= 2 {
			status = "Completed"
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true, "jobId": "J1", "status": status,
		})
	}
}

func TestOrderJobStatus_WatchCtrlCInterruptsSleep(t *testing.T) {
	// Server that never completes, so --watch sits in its 5s default sleep.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true, "jobId": "J1", "status": "InProgress",
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := &cobra.Command{
		Use:           "zr",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return globalflags.Apply(f, cmd)
		},
	}
	globalflags.Register(root)
	grp := &cobra.Command{Use: "order"}
	grp.AddCommand(NewCmdJobStatus(f))
	root.AddCommand(grp)
	root.SetArgs([]string{"order", "job-status", "J1", "--watch"})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel() // simulates Ctrl-C via signal.NotifyContext
	}()

	start := time.Now()
	err := root.ExecuteContext(ctx)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	// The old raw time.Sleep(5s) held cancellation hostage for the full
	// interval — this asserts the sleep is interruptible.
	assert.Less(t, elapsed, 500*time.Millisecond, "Ctrl-C must interrupt the polling sleep promptly")
}

func TestOrderJobStatus_WatchTimeoutGivesUp(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true, "jobId": "J1", "status": "InProgress",
		})
	})

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "job-status", "J1", "--watch", "--interval", "30ms", "--timeout", "80ms")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "gave up waiting for job J1")
	assert.Contains(t, err.Error(), "InProgress")
}

func TestOrderJobStatus_WatchIntervalCompletes(t *testing.T) {
	var calls int32
	handler := inProgressHandler(t, &calls)

	stdout, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "job-status", "J1", "--watch", "--interval", "20ms")

	require.NoError(t, err)
	assert.Contains(t, stdout, "Completed")
	assert.Contains(t, stderr, "polling in 20ms")
	assert.GreaterOrEqual(t, atomic.LoadInt32(&calls), int32(2))
}

func TestOrderJobStatus_WatchRejectsNonPositiveInterval(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "job-status", "J1", "--watch", "--interval", "0s")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--interval must be positive")
}

func TestOrderJobStatus_TimeoutAbortsInFlightRequest(t *testing.T) {
	// Server stalls each request far longer than --timeout; the deadline must
	// abort the in-flight GET, not just the sleep between polls.
	unblock := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-unblock:
		case <-r.Context().Done():
		}
	})
	defer close(unblock)

	start := time.Now()
	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "job-status", "J1", "--watch", "--interval", "20ms", "--timeout", "80ms")
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "gave up waiting for job J1", "mid-request timeout must use the friendly message")
	assert.Less(t, elapsed, 500*time.Millisecond, "--timeout must abort an in-flight request promptly")
}

// TestOrderJobStatus_WatchTreatsCanceledAsTerminal pins the US-spelling fix:
// Zuora emits "Canceled" for async order jobs; --watch must stop, not poll
// forever.
func TestOrderJobStatus_WatchTreatsCanceledAsTerminal(t *testing.T) {
	var calls int32
	handler := func(w http.ResponseWriter, r *http.Request) {
		status := "InProgress"
		if atomic.AddInt32(&calls, 1) >= 2 {
			status = "Canceled"
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success":true,"jobId":"J1","status":%q}`, status)
	}

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "job-status", "J1", "--watch", "--interval", "10ms")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Canceled")
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls), "polling must stop at the terminal status")
}
