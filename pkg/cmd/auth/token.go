package auth

import (
	"fmt"

	iauth "github.com/matsuzj/zuora-cli/internal/auth"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func newCmdToken(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Print the access token",
		Long:  "Print the current access token for use in scripts.\nRefreshes the token if expired.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runToken(f)
		},
	}

	return cmd
}

func runToken(f *factory.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	envName := cfg.ActiveEnvironment()

	creds := iauth.NewCredentialStore()
	ts := &iauth.TokenSource{Config: cfg, Creds: creds}
	token, err := ts.Token(envName)
	if err != nil {
		return err
	}

	fmt.Fprintln(f.IOStreams.Out, token)
	return nil
}
