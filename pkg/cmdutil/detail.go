package cmdutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// Action describes a single detail-view HTTP command: one request, one JSON
// decode, one RenderDetail call, and an optional success message.
type Action struct {
	// Method is the HTTP verb: "GET", "POST", "PUT", "PATCH", "DELETE".
	Method string

	// Path is the API path, e.g. "/v1/invoices/inv-001". Callers must
	// percent-encode dynamic segments with url.PathEscape — RunDetail passes
	// the string verbatim.
	Path string

	// Body is the request body reader; nil sends no body (and therefore no
	// WithBody option).
	Body io.Reader

	// ReqOpts are passed as trailing options after the optional WithBody —
	// api.WithQuery, api.WithHeader, api.WithoutCheckSuccess, etc.
	ReqOpts []api.RequestOption

	// Fields converts the decoded top-level JSON object into the rows
	// RenderDetail shows. Nested responses unwrap inside this closure —
	// RunDetail does not unwrap.
	Fields func(raw map[string]interface{}) []output.DetailField

	// SuccessMsg, when non-nil, runs after a successful render; a non-empty
	// return (with trailing newline) is written to ErrOut, "" suppresses the
	// message (the payment/create pattern: suppress when the id is absent).
	// It receives the same raw map as Fields for value interpolation.
	SuccessMsg func(raw map[string]interface{}) string
}

// RunDetail executes a detail-view command: one HTTP request, one JSON decode,
// one RenderDetail, optional ErrOut success message. It is the shared runner
// for pure detail commands (the P3 migration); variant commands (nested
// unwrap-with-fallback, polling, multi-request, multipart) keep hand-written
// run functions per docs/refactoring-plan.md §P3-1.
func RunDetail(cmd *cobra.Command, f *factory.Factory, act Action) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	reqOpts := make([]api.RequestOption, 0, len(act.ReqOpts)+1)
	if act.Body != nil {
		reqOpts = append(reqOpts, api.WithBody(act.Body))
	}
	reqOpts = append(reqOpts, act.ReqOpts...)

	resp, err := client.Do(act.Method, act.Path, reqOpts...)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	// An empty 2xx body (204, or an action endpoint's empty 200) has no
	// detail to render: report success via the message when one is provided
	// (machine flags get a synthesized {"success":true}); without a message
	// this Action was mis-targeted — deletes with body-less responses belong
	// on RenderDeleteResult, not RunDetail.
	if len(bytes.TrimSpace(resp.Body)) == 0 {
		if act.SuccessMsg != nil {
			return output.RenderSuccess(f.IOStreams, fmtOpts, act.SuccessMsg(map[string]interface{}{}))
		}
		return fmt.Errorf("empty response body (HTTP %d): this command should render via RenderDeleteResult or provide a SuccessMsg", resp.StatusCode)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		// Verbatim error wrap — existing command tests assert this string.
		return fmt.Errorf("parsing response: %w", err)
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, act.Fields(raw)); err != nil {
		return err
	}

	if act.SuccessMsg != nil {
		// Fprint, not Fprintf: the message is pre-formatted, and a dynamic
		// value containing '%' must not be reinterpreted as a format verb.
		if msg := act.SuccessMsg(raw); msg != "" {
			fmt.Fprint(f.IOStreams.ErrOut, msg)
		}
	}
	return nil
}
