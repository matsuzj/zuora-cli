package cmdutil

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/matsuzj/zuora-cli/pkg/output"
)

// RenderDeleteResult renders a delete/void response uniformly. Zuora delete
// endpoints answer in three shapes, which the per-command copies used to
// judge three different ways:
//
//  1. HTTP 204 (no body)            → success
//  2. HTTP 2xx, empty/non-JSON body → success (policy decision in
//     docs/refactoring-plan.md §5: the success-flag check upstream already
//     rejects logical failures, so a bodyless 2xx is a completed delete)
//  3. HTTP 2xx with a JSON body     → render it as a detail view via fields
//
// humanMsg must be a complete sentence with a trailing newline. It goes to
// stderr on every success shape: for shapes 1–2 via RenderSuccess (machine
// flags get a synthesized {"success":true} instead), and AFTER the detail
// render for shape 3 — the same convention as RunDetail.SuccessMsg
// (user-approved delete unification, 2026-06-12). fields builds the detail
// rows for shape 3; pass nil to treat a JSON body like shape 2.
func RenderDeleteResult(ios *iostreams.IOStreams, resp *api.Response, opts output.FormatOptions, humanMsg string, fields func(raw map[string]interface{}) []output.DetailField) error {
	body := bytes.TrimSpace(resp.Body)
	if resp.StatusCode == 204 || len(body) == 0 || !json.Valid(body) || fields == nil {
		return output.RenderSuccess(ios, opts, humanMsg)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	if err := output.RenderDetail(ios, resp.Body, opts, fields(raw)); err != nil {
		return err
	}
	if humanMsg != "" {
		// Fprint, not Fprintf: dynamic values in the message must not be
		// interpreted as format verbs.
		fmt.Fprint(ios.ErrOut, humanMsg)
	}
	return nil
}
