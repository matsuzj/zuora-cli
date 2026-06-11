package cmdutil

import (
	"github.com/spf13/cobra"
)

// AddBodyFlag registers the canonical --body / -b flag. required is accepted
// now so call sites already declare intent; it is wired to MarkFlagRequired in
// the P5 cobra-required migration (docs/refactoring-plan.md) — until then the
// RunE guards keep enforcing it.
func AddBodyFlag(cmd *cobra.Command, dest *string, required bool) {
	cmd.Flags().StringVarP(dest, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")
	_ = required
}

// AddConfirmFlag registers the canonical --confirm flag. operation is the
// noun shown in help ("deletion", "cancellation", "scrub", ...) so the 19
// hand-rolled variants collapse to one definition without hiding the
// per-command wording.
func AddConfirmFlag(cmd *cobra.Command, dest *bool, operation string) {
	cmd.Flags().BoolVar(dest, "confirm", false, "Confirm the "+operation)
}
