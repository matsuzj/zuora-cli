package cmdutil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// detailCmd builds a minimal command carrying the output format flags
// RunDetail resolves via output.FromCmd.
func detailCmd(args ...string) *cobra.Command {
	cmd := &cobra.Command{Use: "x", Run: func(*cobra.Command, []string) {}}
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().String("jq", "", "")
	cmd.Flags().String("template", "", "")
	cmd.Flags().Bool("csv", false, "")
	_ = cmd.Flags().Parse(args)
	return cmd
}

func idFields(raw map[string]interface{}) []output.DetailField {
	return []output.DetailField{{Key: "ID", Value: GetString(raw, "id")}}
}

func TestRunDetail_GETRendersFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/things/T-1", r.URL.Path)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"success":true,"id":"T-1"}`))
	}))
	defer server.Close()

	ios, _, out, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	err := RunDetail(detailCmd(), f, Action{Method: "GET", Path: "/v1/things/T-1", Fields: idFields})
	require.NoError(t, err)
	assert.Contains(t, out.String(), "T-1")
	assert.Empty(t, errOut.String(), "nil SuccessMsg writes nothing")
}

func TestRunDetail_POSTSendsBodyAndSuccessMsg(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		b := make([]byte, 64)
		n, _ := r.Body.Read(b)
		assert.Contains(t, string(b[:n]), `"name":"x"`)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"success":true,"id":"NEW-1"}`))
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	err := RunDetail(detailCmd(), f, Action{
		Method: "POST", Path: "/v1/things", Body: strings.NewReader(`{"name":"x"}`),
		Fields: idFields,
		SuccessMsg: func(raw map[string]interface{}) string {
			if id := GetString(raw, "id"); id != "" {
				return "Thing " + id + " created.\n"
			}
			return ""
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Thing NEW-1 created.\n", errOut.String())
}

func TestRunDetail_SuccessMsgSuppressedWhenEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	err := RunDetail(detailCmd(), f, Action{
		Method: "POST", Path: "/v1/things", Fields: idFields,
		SuccessMsg: func(raw map[string]interface{}) string {
			if id := GetString(raw, "id"); id != "" {
				return "never\n"
			}
			return ""
		},
	})
	require.NoError(t, err)
	assert.Empty(t, errOut.String())
}

func TestRunDetail_ParseErrorWrapVerbatim(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`[1,2,3]`)) // valid JSON, wrong shape for a map
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	err := RunDetail(detailCmd(), f, Action{Method: "GET", Path: "/v1/x", Fields: idFields})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing response: ", "wrap string is load-bearing for existing tests")
}

func TestRunDetail_ReqOptsApplied(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"success":true,"id":"Q"}`))
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	err := RunDetail(detailCmd(), f, Action{
		Method: "GET", Path: "/v1/x",
		ReqOpts: []api.RequestOption{api.WithQuery("page", "2")},
		Fields:  idFields,
	})
	require.NoError(t, err)
}

func TestRunDetail_SuccessFalseErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"success":false,"reasons":[{"code":1,"message":"nope"}]}`))
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	err := RunDetail(detailCmd(), f, Action{Method: "GET", Path: "/v1/x", Fields: idFields})
	require.Error(t, err, "default-on success check applies through the runner")
	assert.Contains(t, err.Error(), "nope")
}

func TestRunDetail_JSONFlagBypassesFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"success":true,"id":"J"}`))
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	err := RunDetail(detailCmd("--json"), f, Action{Method: "GET", Path: "/v1/x", Fields: idFields})
	require.NoError(t, err)
	assert.Contains(t, out.String(), `"id": "J"`)
}

func TestRunDetail_EmptyBodyWithMsgIsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	err := RunDetail(detailCmd(), f, Action{
		Method: "PUT", Path: "/v1/things/T-1/activate", Fields: idFields,
		SuccessMsg: func(map[string]interface{}) string { return "Activated T-1.\n" },
	})
	require.NoError(t, err)
	assert.Equal(t, "Activated T-1.\n", errOut.String())
}

func TestRunDetail_EmptyBodyWithoutMsgIsClearError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	err := RunDetail(detailCmd(), f, Action{Method: "DELETE", Path: "/v1/things/T-1", Fields: idFields})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty response body", "must not surface as 'parsing response: EOF'")
	assert.Contains(t, err.Error(), "RenderDeleteResult")
}
