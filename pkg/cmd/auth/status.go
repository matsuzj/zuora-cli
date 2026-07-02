package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/matsuzj/zuora-cli/internal/auth"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdStatus(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, f)
		},
	}

	return cmd
}

func runStatus(cmd *cobra.Command, f *factory.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	envName := cfg.ActiveEnvironment()

	env, err := cfg.Environment(envName)
	if err != nil {
		return err
	}

	// Credential source. The JSON form uses a stable enum ("keyring"/"env");
	// the human form keeps the descriptive phrase.
	credEnum := "keyring"
	credHuman := "keyring"
	if _, _, ok := auth.EnvCredentials(); ok {
		credEnum = "env"
		credHuman = "environment variables (ZR_CLIENT_ID/ZR_CLIENT_SECRET)"
	}

	token, err := cfg.Token(envName)
	if err != nil {
		return err
	}

	// Derive the token status once so the text, detail, and JSON forms agree.
	var tokenState, tokenHuman, expiresAt string
	switch {
	case token == nil:
		tokenState = "not_authenticated"
		tokenHuman = "not authenticated"
	case token.IsValid():
		tokenState = "valid"
		remaining := time.Until(token.ExpiresAt).Truncate(time.Second)
		tokenHuman = fmt.Sprintf("valid (expires in %s)", remaining)
		expiresAt = token.ExpiresAt.Format(time.RFC3339)
	default:
		tokenState = "expired"
		expiresAt = token.ExpiresAt.Format(time.RFC3339)
		tokenHuman = fmt.Sprintf("expired (expired at %s)", expiresAt)
	}

	// Machine-format flags (--json/--jq/--template/--csv) must be honored, not
	// silently ignored (the local-command output-consistency fix, #453). Follow
	// the version.go pattern: keep the exact human text as the default, and
	// route only the format-flag paths through the shared renderer.
	fmtOpts := output.FromCmd(cmd)
	if fmtOpts.JSON || fmtOpts.JQ != "" || fmtOpts.Template != "" || fmtOpts.CSV {
		data := map[string]interface{}{
			"environment": envName,
			"baseUrl":     env.BaseURL,
			"credentials": credEnum,
			"token": map[string]interface{}{
				"status":    tokenState,
				"expiresAt": expiresAt,
			},
		}
		rawJSON, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshaling status: %w", err)
		}
		fields := []output.DetailField{
			{Key: "Environment", Value: envName},
			{Key: "Base URL", Value: env.BaseURL},
			{Key: "Credentials", Value: credHuman},
			{Key: "Token", Value: tokenHuman},
		}
		return output.RenderDetail(f.IOStreams, rawJSON, fmtOpts, fields)
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "Environment: %s\n", envName)
	fmt.Fprintf(out, "Base URL:    %s\n", env.BaseURL)
	fmt.Fprintf(out, "Credentials: %s\n", credHuman)
	if token == nil {
		fmt.Fprintln(out, "Token:       not authenticated")
		return nil
	}
	fmt.Fprintf(out, "Token:       %s\n", tokenHuman)
	return nil
}
