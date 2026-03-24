package auth

import (
	"fmt"
	"os"

	iauth "github.com/matsuzj/zuora-cli/internal/auth"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func newCmdLogout(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove authentication credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(f)
		},
	}

	return cmd
}

func runLogout(f *factory.Factory) error {
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

	fmt.Fprintf(f.IOStreams.Out, "Logged out of %s\n", envName)

	// Warn if env vars will still provide credentials
	if os.Getenv("ZR_CLIENT_ID") != "" && os.Getenv("ZR_CLIENT_SECRET") != "" {
		fmt.Fprintln(f.IOStreams.ErrOut, "Note: ZR_CLIENT_ID/ZR_CLIENT_SECRET environment variables are still set.")
		fmt.Fprintln(f.IOStreams.ErrOut, "Unset them to fully disable authentication.")
	}

	return nil
}
