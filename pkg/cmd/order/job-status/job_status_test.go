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

// completedCreateOrderJob is the live-verified GET /v1/async-jobs/{jobId}
// response for a completed AsyncCreateOrder job (probed 2026-07-02). The
// identifying fields (orderNumber, accountNumber, subscriptionNumber) live
// NESTED under `result`, and there is deliberately no root jobId/orderNumber/
// accountNumber — that absence is what the fields fix depends on, so the
// fixture must not reintroduce them.
func completedCreateOrderJob() map[string]interface{} {
	return map[string]interface{}{
		"status": "Completed",
		"errors": nil,
		"result": map[string]interface{}{
			"orderNumber":   "O-00014288",
			"accountNumber": "A00023286",
			"status":        "Completed",
			"subscriptions": []interface{}{
				map[string]interface{}{
					"subscriptionNumber": "A-S00005866",
					"status":             "有効",
				},
			},
			"jobType": "AsyncCreateOrder",
		},
		"success": true,
	}
}

func TestOrderJobStatus_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/async-jobs/2c92c0f9876", completedCreateOrderJob())

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "job-status", "2c92c0f9876")
	require.NoError(t, err)
	assert.Contains(t, stdout, "2c92c0f9876", "Job ID row echoes the queried argument")
	assert.Contains(t, stdout, "Completed")
	// Sourced from result.orderNumber / result.accountNumber (nested), not root.
	assert.Contains(t, stdout, "O-00014288")
	assert.Contains(t, stdout, "A00023286")
	// Sourced from result.subscriptions[0].subscriptionNumber.
	assert.Contains(t, stdout, "A-S00005866")
}

// TestOrderJobStatus_ResultObjectNotRenderedRaw pins the object-result fix:
// `result` is an object, and the old code read it via GetString, dumping a
// Go-map representation ("map[...]") into the detail view. This asserts no such
// dump leaks — reverting to GetString(raw,"result") makes it fail.
func TestOrderJobStatus_ResultObjectNotRenderedRaw(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/async-jobs/2c92c0f9876", completedCreateOrderJob())

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "job-status", "2c92c0f9876")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "map[", "result object must be descended into, never GetString'd into a Go-map dump")
}

// TestOrderJobStatus_RendersErrors pins the errors row for a failed job. The
// failed-job error shape ({code, message}) is Zuora's documented convention but
// was not live-verified (all probed jobs completed); formatJobErrors is
// defensive so an unexpected shape yields a blank row, never a Go-map dump.
func TestOrderJobStatus_RendersErrors(t *testing.T) {
	// A failed JOB still returns a successful READ (root success:true; the
	// failure is conveyed by status + errors). Using success:false here would
	// instead trip the client's success-flag gate before rendering.
	handler := cmdtest.OK(t, "GET", "/v1/async-jobs/job-fail", map[string]interface{}{
		"status": "Failed",
		"errors": []interface{}{
			map[string]interface{}{"code": "58730020", "message": "Invalid rate plan"},
		},
		"result":  nil,
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "job-status", "job-fail")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Failed")
	assert.Contains(t, stdout, "58730020: Invalid rate plan")
	assert.NotContains(t, stdout, "map[", "errors array must be formatted, never dumped as a Go map")
}

func TestOrderJobStatus_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "job-status")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
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
	// 2s bound: generous for loaded CI runners, still far below the 5s default
	// interval this test exists to prove is interruptible.
	assert.Less(t, elapsed, 2*time.Second, "Ctrl-C must interrupt the polling sleep promptly")
}

func TestOrderJobStatus_WatchTimeoutGivesUp(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true, "jobId": "J1", "status": "InProgress",
		})
	})

	// 400ms timeout with a 20ms interval: at least one poll reliably completes
	// before the deadline even on a loaded runner, so the "last status" in the
	// give-up message is deterministically InProgress (80ms left only ~2 polls
	// of headroom and flaked if the first poll ran slow).
	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "job-status", "J1", "--watch", "--interval", "20ms", "--wait-timeout", "400ms")

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
	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "job-status", "J1", "--watch", "--interval", "20ms", "--wait-timeout", "80ms")
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "gave up waiting for job J1", "mid-request timeout must use the friendly message")
	// 2s bound: the server stalls forever, so anything under 2s proves the
	// deadline aborted the in-flight GET (500ms flirted with CI scheduling noise).
	assert.Less(t, elapsed, 2*time.Second, "--timeout must abort an in-flight request promptly")
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

// TestOrderJobStatus_WatchPollLineSanitized pins that the --watch progress
// line sanitizes the response-derived status: a hostile value must not write
// escape codes to the terminal via stderr on every poll.
func TestOrderJobStatus_WatchPollLineSanitized(t *testing.T) {
	handler := cmdtest.Sequence(
		cmdtest.OK(t, "", "", map[string]interface{}{"status": "In Progress\x1b[2J\x1b[H", "success": true}),
		cmdtest.OK(t, "", "", map[string]interface{}{"status": "Completed", "success": true}),
	)

	_, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "job-status", "J1", "--watch", "--interval", "5ms")
	require.NoError(t, err)
	assert.Contains(t, stderr, "polling in")
	assert.NotContains(t, stderr, "\x1b", "response-derived status must be sanitized on the poll line")
}

// TestOrderJobStatus_TimeoutAliasRemoved pins the #512 removal: the old
// local --timeout alias is gone, so the spelling now falls through to the
// GLOBAL persistent --timeout — under cmdtest's root that flag exists, so
// the local wait keeps running and the global deadline aborts the command
// with the raw deadline error, not the local friendly give-up message.
func TestOrderJobStatus_TimeoutAliasRemoved(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true, "jobId": "J1", "status": "InProgress",
		})
	})

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "job-status", "J1", "--watch", "--interval", "20ms", "--timeout", "400ms")
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "gave up waiting for job J1",
		"the local alias must be gone; --timeout now means the GLOBAL deadline")
}
