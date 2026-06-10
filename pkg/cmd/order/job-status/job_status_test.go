package jobstatus

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().String("jq", "", "")
	root.PersistentFlags().String("template", "", "")
	order := &cobra.Command{Use: "order"}
	order.AddCommand(NewCmdJobStatus(f))
	root.AddCommand(order)
	return root
}

func TestOrderJobStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/async-jobs/2c92c0f9876", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       true,
			"jobId":         "2c92c0f9876",
			"status":        "Completed",
			"result":        "Success",
			"orderNumber":   "O-00000001",
			"accountNumber": "A001",
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "job-status", "2c92c0f9876"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "2c92c0f9876")
	assert.Contains(t, out.String(), "Completed")
	assert.Contains(t, out.String(), "O-00000001")
}

func TestOrderJobStatus_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "job-status"})
	err := root.Execute()

	assert.Error(t, err)
}

// Always-InProgress server for watch tests.
func inProgressServer(t *testing.T, calls *int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(calls, 1)
		w.WriteHeader(200)
		status := "InProgress"
		if n >= 2 {
			status = "Completed"
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true, "jobId": "J1", "status": status,
		})
	}))
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

	root := newTestRoot(f)
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true, "jobId": "J1", "status": "InProgress",
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "job-status", "J1", "--watch", "--interval", "30ms", "--timeout", "80ms"})

	start := time.Now()
	err := root.Execute()
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "gave up waiting for job J1")
	assert.Contains(t, err.Error(), "InProgress")
	assert.Less(t, elapsed, 500*time.Millisecond)
}

func TestOrderJobStatus_WatchIntervalCompletes(t *testing.T) {
	var calls int32
	server := inProgressServer(t, &calls)
	defer server.Close()

	ios, _, out, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "job-status", "J1", "--watch", "--interval", "20ms"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Completed")
	assert.Contains(t, errOut.String(), "polling in 20ms")
	assert.GreaterOrEqual(t, atomic.LoadInt32(&calls), int32(2))
}
