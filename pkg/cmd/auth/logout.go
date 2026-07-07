package auth

import (
	"fmt"

	iauth "github.com/matsuzj/zuora-cli/internal/auth"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdLogout(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove authentication credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(cmd, f)
		},
	}

	return cmd
}

func runLogout(cmd *cobra.Command, f *factory.Factory) error {
	fmtOpts := output.FromCmd(cmd)
	// Reject bare --csv BEFORE mutating credential state (write-command contract).
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

	cfg, err := f.Config()
	if err != nil {
		return err
	}

	envName := cfg.ActiveEnvironment()

	// Always attempt to delete from keyring directly (not env var store)
	if err := iauth.KeyringStore().Delete(envName); err != nil {
		fmt.Fprintf(f.IOStreams.ErrOut, "Warning: could not remove keyring credentials: %s\n", err)
	}

	if err := cfg.RemoveToken(envName); err != nil {
		return err
	}
	if err := cfg.Save(); err != nil {
		return err
	}

	// Warn if env vars will still provide credentials
	if _, _, ok := iauth.EnvCredentials(); ok {
		fmt.Fprintln(f.IOStreams.ErrOut, "Note: ZR_CLIENT_ID/ZR_CLIENT_SECRET environment variables are still set.")
		fmt.Fprintln(f.IOStreams.ErrOut, "Unset them to fully disable authentication.")
	}

	// Machine flags get {"success": true}; the human message goes to stderr,
	// keeping stdout clean (#453/#519).
	return output.RenderSuccess(f.IOStreams, fmtOpts, fmt.Sprintf("Logged out of %s\n", envName))
}
