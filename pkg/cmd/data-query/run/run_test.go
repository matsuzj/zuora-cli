package run

import (
	"context"
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

	handler := cmdtest.Route(t, map[string]http.HandlerFunc{
		"/query/jobs": cmdtest.OK(t, "POST", "", map[string]interface{}{
			"data": map[string]interface{}{"id": "job-1", "queryStatus": "accepted"},
		}),
		"/query/jobs/job-1": cmdtest.Sequence(
			cmdtest.OK(t, "GET", "", map[string]interface{}{
				"data": map[string]interface{}{"id": "job-1", "queryStatus": "in_progress", "outputRows": "1"},
			}),
			cmdtest.OK(t, "GET", "", map[string]interface{}{
				"data": map[string]interface{}{"id": "job-1", "queryStatus": "completed", "outputRows": "1", "dataFile": resultSrv.URL + "/result"},
			}),
		),
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
			w.Write([]byte(`{"data":{"id":"job-1","queryStatus":"completed","outputRows":"7","processingTime":321,"dataFile":"https://dq.example.invalid/files/job-1.json"}}`))
			return
		}
		t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
	})
	stdout, _, err := cmdtest.Run(t, "data-query", newCmd, handler, "data-query", "run", "SELECT 1", "--interval", "5ms", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, "completed")
}

// TestRun_NoOutputDetailFieldsLabelBound pins the no---output detail rendering
// (dqutil.DetailFields): the fixture carries EVERY key the renderer reads with
// a distinctive value, and each field is asserted under its own label, so a
// key typo or nesting mistake renders "" and fails here (fixture-masking,
// #482). No --output means the dataFile URL is never fetched.
func TestRun_NoOutputDetailFieldsLabelBound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(200)
			w.Write([]byte(`{"data":{"id":"job-detail-1","queryStatus":"completed","outputRows":4321,"processingTime":987,"dataFile":"https://dq.example.invalid/files/res-42.json"}}`))
			return
		}
		t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
	})
	stdout, _, err := cmdtest.Run(t, "data-query", newCmd, handler, "data-query", "run", "SELECT 1", "--interval", "5ms")
	require.NoError(t, err)
	assert.Regexp(t, `(?m)^ID:\s+job-detail-1$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+completed$`, stdout)
	// JSON numbers render as plain decimals via GetDecimal (4321, not 4.321e+03).
	assert.Regexp(t, `(?m)^Output Rows:\s+4321$`, stdout)
	assert.Regexp(t, `(?m)^Processing Time:\s+987$`, stdout)
	assert.Regexp(t, `(?m)^Data File:\s+https://dq\.example\.invalid/files/res-42\.json$`, stdout)
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
	// 150ms deadline: ample room for the submit POST to complete first on a
	// loaded runner, so the deadline deterministically lands in the poll loop
	// and produces the friendly give-up message (30ms could expire mid-submit).
	_, _, err := cmdtest.Run(t, "data-query", newCmd, handler, "data-query", "run", "SELECT 1", "--interval", "5ms", "--timeout", "150ms")
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

// TestRun_PollLineSanitized pins that the poll progress line sanitizes the
// response-derived job id/status: hostile values must not write escape codes
// to the terminal via stderr.
func TestRun_PollLineSanitized(t *testing.T) {
	handler := cmdtest.Route(t, map[string]http.HandlerFunc{
		"/query/jobs": cmdtest.OK(t, "POST", "", map[string]interface{}{
			"data": map[string]interface{}{"id": "job-1", "queryStatus": "accepted\x1b[2J\x1b[H"},
		}),
		"/query/jobs/job-1": cmdtest.Sequence(
			cmdtest.OK(t, "GET", "", map[string]interface{}{
				"data": map[string]interface{}{"id": "job-1", "queryStatus": "in_progress\x1b[31m", "outputRows": "3"},
			}),
			cmdtest.OK(t, "GET", "", map[string]interface{}{
				"data": map[string]interface{}{"id": "job-1", "queryStatus": "completed", "outputRows": "3"},
			}),
		),
	})

	_, stderr, err := cmdtest.Run(t, "data-query", newCmd, handler, "data-query", "run", "SELECT 1", "--interval", "5ms")
	require.NoError(t, err)
	assert.Contains(t, stderr, "polling in")
	assert.NotContains(t, stderr, "\x1b", "response-derived values must be sanitized on the poll line")
}
