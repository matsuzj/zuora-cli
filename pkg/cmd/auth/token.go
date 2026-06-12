package auth

import (
	"fmt"
	"os"

	iauth "github.com/matsuzj/zuora-cli/internal/auth"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/globalflags"
	"github.com/spf13/cobra"
)

func newCmdToken(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Print the access token",
		Long:  "Print the current access token for use in scripts.\nRefreshes the token if expired.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runToken(cmd, f)
		},
	}

	return cmd
}

func runToken(cmd *cobra.Command, f *factory.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	envName := cfg.ActiveEnvironment()

	creds := iauth.NewCredentialStore()
	ts := &iauth.TokenSource{Config: cfg, Creds: creds}
	// Same gate as the factory path: ZR_DEBUG=api implies verbose, so the
	// hand-built TokenSource here must not lose the auth lines (Codex).
	vCount, _ := cmd.Flags().GetCount("verbose")
	if verbose, _ := globalflags.VerboseLevels(vCount, os.Getenv("ZR_DEBUG")); verbose {
		ts.Logf = func(format string, args ...any) { fmt.Fprintf(f.IOStreams.ErrOut, format, args...) }
	}
	// TokenContext so a hung OAuth endpoint is interruptible with Ctrl-C.
	token, err := ts.TokenContext(cmd.Context(), envName)
	if err != nil {
		return err
	}

	fmt.Fprintln(f.IOStreams.Out, token)
	return nil
}
