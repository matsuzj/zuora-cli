package run

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/data-query/dqutil"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdRun(f) }

func hardenedClientTrusting(srv *httptest.Server) *http.Client {
	hc := dqutil.HardenedDownloadClient()
	hc.Transport.(*http.Transport).TLSClientConfig = srv.Client().Transport.(*http.Transport).TLSClientConfig
	return hc
}

func TestRun_PollsThenDownloads(t *testing.T) {
	resultSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Authorization"), "result download must not carry Authorization")
		w.Write([]byte("col\n1\n"))
	}))
	defer resultSrv.Close()

	dqutil.DownloadClientForTest = hardenedClientTrusting(resultSrv)
	defer func() { dqutil.DownloadClientForTest = nil }()

	polls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/query/jobs":
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"id": "job-1", "queryStatus": "accepted"}})
		case r.Method == "GET" && r.URL.Path == "/query/jobs/job-1":
			polls++
			d := map[string]interface{}{"id": "job-1", "queryStatus": "in_progress", "outputRows": "1"}
			if polls >= 2 {
				d["queryStatus"] = "completed"
				d["dataFile"] = resultSrv.URL + "/result"
			}
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(map[string]interface{}{"data": d})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	out := filepath.Join(t.TempDir(), "res.csv")
	_, stderr, err := cmdtest.Run(t, "data-query", newCmd, handler, "data-query", "run", "SELECT 1", "--interval", "5ms", "--output", out)
	require.NoError(t, err)

	b, rerr := os.ReadFile(out)
	require.NoError(t, rerr)
	assert.Equal(t, "col\n1\n", string(b))
	assert.Contains(t, stderr, "completed")
}

func TestRun_NoOutputRendersMetadata(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(200)
			w.Write([]byte(`{"data":{"id":"job-1","queryStatus":"completed","outputRows":"7"}}`))
			return
		}
		t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
	})
	stdout, _, err := cmdtest.Run(t, "data-query", newCmd, handler, "data-query", "run", "SELECT 1", "--interval", "5ms", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, "completed")
}

func TestRun_FailedJob(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(200)
			w.Write([]byte(`{"data":{"id":"job-1","queryStatus":"failed","errorMessage":"bad sql"}}`))
			return
		}
		t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
	})
	_, _, err := cmdtest.Run(t, "data-query", newCmd, handler, "data-query", "run", "SELECT bad", "--interval", "5ms")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
	assert.Contains(t, err.Error(), "bad sql")
}

func TestRun_TimeoutWhileQueued(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"data":{"id":"job-1","queryStatus":"accepted"}}`))
	})
	_, _, err := cmdtest.Run(t, "data-query", newCmd, handler, "data-query", "run", "SELECT 1", "--interval", "5ms", "--timeout", "30ms")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gave up waiting")
}

func TestRun_GlobalTimeoutNoMisleadingMessage(t *testing.T) {
	// When the GLOBAL `zr --timeout` deadline (carried on the command context)
	// fires during the poll sleep with the LOCAL --timeout unset (opts.Timeout
	// == 0), the error must NOT be the "gave up ... after 0s ... raise --timeout"
	// message — that duration is meaningless and the hint points at the wrong
	// knob. Bites if the sleep-path waitErr loses its opts.Timeout>0 guard. (#428)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Submit succeeds; the job never reaches a terminal state.
		w.WriteHeader(200)
		w.Write([]byte(`{"data":{"id":"job-1","queryStatus":"accepted"}}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), srv.URL, "tok")

	cmd := NewCmdRun(f)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	// Simulate the global deadline; --interval (200ms) is far longer so the
	// deadline lands during the first SleepContext, not a request.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"SELECT 1", "--interval", "200ms"}) // no local --timeout

	err := cmd.Execute()
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "after 0s")
	assert.NotContains(t, err.Error(), "raise --timeout")
}
