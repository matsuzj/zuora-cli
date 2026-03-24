package auth

import (
	"fmt"
	"os"

	"github.com/matsuzj/zuora-cli/internal/auth"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type loginOptions struct {
	Factory      *factory.Factory
	ClientID     string
	ClientSecret string
}

func newCmdLogin(f *factory.Factory) *cobra.Command {
	opts := &loginOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Zuora",
		Long: `Authenticate with a Zuora environment using OAuth 2.0 client credentials.

Credentials can be provided via flags, environment variables (ZR_CLIENT_ID,
ZR_CLIENT_SECRET), or interactive prompts.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(opts)
		},
	}

	cmd.Flags().StringVar(&opts.ClientID, "client-id", "", "OAuth client ID")
	cmd.Flags().StringVar(&opts.ClientSecret, "client-secret", "", "OAuth client secret")
	// NOTE: --env is inherited from root persistent flags, not defined locally

	return cmd
}

func runLogin(opts *loginOptions) error {
	f := opts.Factory
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	// --env is handled by root PersistentPreRunE which overrides cfg.ActiveEnvironment()
	envName := cfg.ActiveEnvironment()

	env, err := cfg.Environment(envName)
	if err != nil {
		return err
	}

	clientID := opts.ClientID
	clientSecret := opts.ClientSecret

	// Fall back to environment variables
	if clientID == "" {
		clientID = os.Getenv("ZR_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("ZR_CLIENT_SECRET")
	}

	// Interactive prompts if still not provided
	if clientID == "" || clientSecret == "" {
		ios := f.IOStreams
		// Check stdin is a terminal (not stdout) for interactive detection
		stdinIsTTY := false
		if f, ok := ios.In.(interface{ Fd() uintptr }); ok {
			stdinIsTTY = term.IsTerminal(int(f.Fd()))
		}
		if !stdinIsTTY {
			return fmt.Errorf("client-id and client-secret flags (or ZR_CLIENT_ID/ZR_CLIENT_SECRET env vars) are required in non-interactive mode")
		}

		if clientID == "" {
			fmt.Fprint(ios.ErrOut, "Client ID: ")
			var id string
			if _, err := fmt.Fscan(ios.In, &id); err != nil {
				return fmt.Errorf("reading client ID: %w", err)
			}
			clientID = id
		}

		if clientSecret == "" {
			fmt.Fprint(ios.ErrOut, "Client Secret: ")
			type fder interface{ Fd() uintptr }
			if f, ok := ios.In.(fder); ok {
				secret, err := term.ReadPassword(int(f.Fd()))
				if err != nil {
					return fmt.Errorf("reading client secret: %w", err)
				}
				fmt.Fprintln(ios.ErrOut)
				clientSecret = string(secret)
			} else {
				var secret string
				if _, err := fmt.Fscan(ios.In, &secret); err != nil {
					return fmt.Errorf("reading client secret: %w", err)
				}
				clientSecret = secret
			}
		}
	}

	// Validate credentials by fetching a token first (don't persist invalid credentials)
	creds := &auth.MockCredentialStore{Creds: map[string][2]string{
		envName: {clientID, clientSecret},
	}}
	ts := &auth.TokenSource{Config: cfg, Creds: creds}
	_, err = ts.Refresh(envName)
	if err != nil {
		return err
	}

	// Credentials are valid — now persist to keyring
	if err := auth.KeyringStore().Set(envName, clientID, clientSecret); err != nil {
		fmt.Fprintf(f.IOStreams.ErrOut, "Warning: could not store credentials in keyring: %s\n", err)
		fmt.Fprintln(f.IOStreams.ErrOut, "Token will be cached but credentials are not persisted.")
		fmt.Fprintln(f.IOStreams.ErrOut, "Set ZR_CLIENT_ID and ZR_CLIENT_SECRET environment variables for persistent access.")
	}

	fmt.Fprintf(f.IOStreams.Out, "Logged in to %s (%s)\n", envName, env.BaseURL)
	return nil
}
