package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func newCmdStatus(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(f)
		},
	}

	return cmd
}

func runStatus(f *factory.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	envName := cfg.ActiveEnvironment()

	env, err := cfg.Environment(envName)
	if err != nil {
		return err
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "Environment: %s\n", envName)
	fmt.Fprintf(out, "Base URL:    %s\n", env.BaseURL)

	// Credential source
	credSource := "keyring"
	if os.Getenv("ZR_CLIENT_ID") != "" && os.Getenv("ZR_CLIENT_SECRET") != "" {
		credSource = "environment variables (ZR_CLIENT_ID/ZR_CLIENT_SECRET)"
	}
	fmt.Fprintf(out, "Credentials: %s\n", credSource)

	// Token status
	token, err := cfg.Token(envName)
	if err != nil {
		return err
	}
	if token == nil {
		fmt.Fprintln(out, "Token:       not authenticated")
		return nil
	}

	if token.IsValid() {
		remaining := time.Until(token.ExpiresAt).Truncate(time.Second)
		fmt.Fprintf(out, "Token:       valid (expires in %s)\n", remaining)
	} else {
		fmt.Fprintf(out, "Token:       expired (expired at %s)\n", token.ExpiresAt.Format(time.RFC3339))
	}

	return nil
}
