package cmdtest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newProbeCmd is a minimal but realistic command: GET /v1/probe/<id>, render a
// detail view — enough to exercise the server wiring, persistent flags, and
// both streams through the harness.
func newProbeCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:  "probe <id>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HttpClient()
			if err != nil {
				return err
			}
			resp, err := client.Get("/v1/probe/" + args[0])
			if err != nil {
				return err
			}
			var raw map[string]interface{}
			if err := json.Unmarshal(resp.Body, &raw); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			fields := []output.DetailField{{Key: "Name", Value: cmdutil.GetString(raw, "name")}}
			return output.RenderDetail(f.IOStreams, resp.Body, output.FromCmd(cmd), fields)
		},
	}
}

func TestRun_RootLevelCommand(t *testing.T) {
	stdout, stderr, err := Run(t, "", newProbeCmd,
		OK(t, "GET", "/v1/probe/P-1", map[string]interface{}{"success": true, "name": "Widget"}),
		"probe", "P-1")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Widget")
	assert.Empty(t, stderr)
}

func TestRun_UnderParentGroup(t *testing.T) {
	stdout, _, err := Run(t, "thing", newProbeCmd,
		OK(t, "GET", "/v1/probe/P-2", map[string]interface{}{"success": true, "name": "Gadget"}),
		"thing", "probe", "P-2")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Gadget")
}

func TestRun_PersistentFlagsWired(t *testing.T) {
	stdout, _, err := Run(t, "", newProbeCmd,
		OK(t, "", "", map[string]interface{}{"success": true, "name": "Widget"}),
		"probe", "P-1", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"name": "Widget"`, "--json must reach output.FromCmd through the stub root")
}

func TestRun_ReasonsEnvelopeErrors(t *testing.T) {
	// The success-flag check is on by default, so the Reasons envelope must
	// surface as a non-zero error carrying the message.
	_, _, err := Run(t, "", newProbeCmd, Reasons(t, 53100020, "Missing field"), "probe", "P-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Missing field")
}

func TestRun_ObjectCRUDFailureEnvelopeErrors(t *testing.T) {
	// The default success-flag check covers the uppercase Object-CRUD envelope
	// too, so {"Success":false,"Errors":[...]} must surface as a non-zero error
	// carrying the message. (The client currently renders this shape via its
	// raw-body fallback — see api.parseAPIError — so the message is still present
	// even though the Code/Message are not yet pulled out cleanly.)
	_, _, err := Run(t, "", newProbeCmd, ObjectCRUDFailure(t, "INVALID_VALUE", "Missing field"), "probe", "P-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Missing field")
}

func TestRun_NilHandlerForValidationTests(t *testing.T) {
	_, _, err := Run(t, "", newProbeCmd, nil, "probe")
	require.Error(t, err, "arg validation fails before any HTTP call")
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}

func TestStatus_ExplicitCode(t *testing.T) {
	_, _, err := Run(t, "", newProbeCmd,
		Status(t, "GET", "/v1/probe/P-9", 404, map[string]interface{}{"message": "not found"}),
		"probe", "P-9")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestRoute_DispatchesByPath(t *testing.T) {
	// Two endpoints registered; the probe command hits /v1/probe/P-1 and must be
	// routed to THAT handler (proving Route dispatches by path, not first-match).
	routes := map[string]http.HandlerFunc{
		"/v1/probe/P-1": OK(t, "GET", "/v1/probe/P-1", map[string]interface{}{"success": true, "name": "Widget"}),
		"/v1/probe/P-2": OK(t, "GET", "/v1/probe/P-2", map[string]interface{}{"success": true, "name": "Gadget"}),
	}
	stdout, _, err := Run(t, "", newProbeCmd, Route(t, routes), "probe", "P-1")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Widget")
	assert.NotContains(t, stdout, "Gadget", "the P-2 route must not fire for a P-1 request")
}

func TestExpect_AssertsRequestAndResponds(t *testing.T) {
	// Drive the handler directly with a fully-matching request: method, path,
	// query, header, and JSON body are all asserted, then Respond is returned.
	h := Expect{
		Method:   "POST",
		Path:     "/v1/orders",
		Query:    map[string]string{"async": "true"},
		Headers:  map[string]string{"Content-Type": "application/json"},
		JSONBody: `{"existingAccountNumber":"A001"}`,
		Respond:  map[string]interface{}{"success": true, "orderNumber": "O-1"},
	}.Handler(t)

	req := httptest.NewRequest("POST", "/v1/orders?async=true", strings.NewReader(`{"existingAccountNumber":"A001"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h(rec, req)

	assert.Equal(t, 200, rec.Code)
	assert.Contains(t, rec.Body.String(), "O-1")
}

func TestRun_WithExpect(t *testing.T) {
	// Expect plugs into Run like any handler; the reached-guard passes because
	// the probe command makes the GET.
	stdout, _, err := Run(t, "", newProbeCmd,
		Expect{Method: "GET", Path: "/v1/probe/P-1", Respond: map[string]interface{}{"success": true, "name": "Widget"}}.Handler(t),
		"probe", "P-1")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Widget")
}

func TestLoadFixture(t *testing.T) {
	b := LoadFixture(t, "order_get")
	// Nested, drift-prone keys are present in the captured shape.
	assert.Contains(t, string(b), "O-00000001")
	assert.Contains(t, string(b), "ACCT-9000001")
}

func TestSequence_DispatchesInOrderAndRepeatsLast(t *testing.T) {
	h := Sequence(
		OK(t, "", "", map[string]interface{}{"n": 1}),
		OK(t, "", "", map[string]interface{}{"n": 2}),
	)
	// Three requests over two handlers: 1, 2, then the last repeats.
	for _, want := range []string{`"n":1`, `"n":2`, `"n":2`} {
		rec := httptest.NewRecorder()
		h(rec, httptest.NewRequest("GET", "/poll", nil))
		body := strings.ReplaceAll(rec.Body.String(), " ", "")
		assert.Contains(t, body, want)
	}
}

func TestZOQLPages_ServesQueryThenQueryMore(t *testing.T) {
	h := ZOQLPages(t,
		[]map[string]interface{}{{"Id": "001"}},
		[]map[string]interface{}{{"Id": "002"}},
	)

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest("POST", "/v1/action/query", strings.NewReader(`{"queryString":"SELECT Id FROM Account"}`)))
	var page1 map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &page1))
	assert.Equal(t, false, page1["done"], "page 1 must be done:false")
	locator, _ := page1["queryLocator"].(string)
	require.NotEmpty(t, locator, "page 1 must carry a queryLocator")

	rec = httptest.NewRecorder()
	h(rec, httptest.NewRequest("POST", "/v1/action/queryMore",
		strings.NewReader(fmt.Sprintf(`{"queryLocator":%q}`, locator))))
	var page2 map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &page2))
	assert.Equal(t, true, page2["done"], "page 2 must be done:true")
	assert.Contains(t, rec.Body.String(), "002")
	// The exactly-two-calls cleanup passes because both pages were fetched.
}

// newConfirmProbeCmd guards a (never-reached) write behind the canonical
// cmdutil.RequireConfirm gate — the shape RequiresConfirm is built to pin.
func newConfirmProbeCmd(f *factory.Factory) *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:  "cprobe",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(confirm); err != nil {
				return err
			}
			client, err := f.HttpClient()
			if err != nil {
				return err
			}
			_, err = client.Post("/v1/probe", nil)
			return err
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm")
	return cmd
}

func TestRequiresConfirm_PinsTheGuard(t *testing.T) {
	RequiresConfirm(t, "", newConfirmProbeCmd, "cprobe")
}

// newWriteProbeCmd POSTs — for asserting the harness applies real global-flag
// behavior (--read-only must block it before any HTTP call).
func newWriteProbeCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:  "wprobe",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HttpClient()
			if err != nil {
				return err
			}
			_, err = client.Post("/v1/probe", nil)
			return err
		},
	}
}

func TestRun_AppliesRealGlobalFlagBehavior(t *testing.T) {
	t.Run("--json and --template are rejected", func(t *testing.T) {
		_, _, err := Run(t, "", newProbeCmd, nil, "probe", "P-1", "--json", "--template", "{{.}}")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot use --json and --template together")
	})

	t.Run("--read-only blocks a write before any HTTP call", func(t *testing.T) {
		_, _, err := Run(t, "", newWriteProbeCmd, nil, "wprobe", "--read-only")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read-only")
	})

	// NOTE: --env validation is part of Apply but cannot be asserted here:
	// NewTestFactory pre-wires HttpClient to the test server and never
	// consults f.Config, so the override (and its validation) is bypassed —
	// the same limitation every existing command test has.
}
