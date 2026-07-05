package globalflags_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/globalflags"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --timeout bounds the whole command: a request that would take longer than
// the flag is aborted with context.DeadlineExceeded (which main maps to exit 1,
// NOT the Ctrl-C context.Canceled → exit 130 path).
func TestApply_TimeoutFlagAbortsSlowRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond after 600ms, OR return early when the client cancels (the
		// --timeout deadline). With a 20ms timeout the deadline wins; with NO
		// timeout the 600ms response wins — which is what makes the test bite.
		// 600ms (was 200ms) keeps a wide margin over the 20ms deadline so a
		// loaded CI runner cannot let the response beat a late-firing timer.
		select {
		case <-time.After(600 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), srv.URL, "tok")

	cmd := &cobra.Command{Use: "x"}
	globalflags.Register(cmd)
	require.NoError(t, cmd.ParseFlags([]string{"--timeout", "20ms"}))
	require.NoError(t, globalflags.Apply(f, cmd))

	client, err := f.HttpClient()
	require.NoError(t, err)
	_, err = client.Get("/v1/slow")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded, "the --timeout deadline must surface as DeadlineExceeded")
	assert.False(t, errors.Is(err, context.Canceled),
		"a deadline is NOT a cancellation, so main maps it to exit 1, not 130")
}

// Without --timeout the command is not bounded by a deadline (the slow request
// completes). This is the other side of the bite: it confirms the deadline only
// applies when the flag is set.
func TestApply_NoTimeoutLeavesRequestUnbounded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(30 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), srv.URL, "tok")

	cmd := &cobra.Command{Use: "x"}
	globalflags.Register(cmd)
	require.NoError(t, cmd.ParseFlags(nil))
	require.NoError(t, globalflags.Apply(f, cmd))

	client, err := f.HttpClient()
	require.NoError(t, err)
	_, err = client.Get("/v1/slow")
	require.NoError(t, err, "with no --timeout the 30ms request should complete")
}

// On a subcommand that defines its OWN local --timeout (like order job-status),
// the global `zr --timeout … <cmd>` must still apply: Apply reads the ROOT
// persistent flag, not the locally-shadowed value in the merged flag set.
func TestApply_TimeoutReadsRootDespiteLocalShadow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(600 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), srv.URL, "tok")

	root := &cobra.Command{Use: "zr"}
	globalflags.Register(root)
	child := &cobra.Command{Use: "job-status"}
	child.Flags().Duration("timeout", 0, "local watch deadline (shadows the global)")
	root.AddCommand(child)

	// `zr --timeout 20ms job-status`: the GLOBAL flag is set on the root.
	require.NoError(t, root.PersistentFlags().Set("timeout", "20ms"))
	require.NoError(t, globalflags.Apply(f, child))

	client, err := f.HttpClient()
	require.NoError(t, err)
	_, err = client.Get("/v1/slow")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded,
		"the global --timeout must apply even when the subcommand shadows the flag name")
}
