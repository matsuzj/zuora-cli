// Package post implements the "zr usage post" command.
package post

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type postOptions struct {
	Factory *factory.Factory
	File    string
}

// NewCmdPost creates the usage post command.
func NewCmdPost(f *factory.Factory) *cobra.Command {
	opts := &postOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "post",
		Short: "Upload a usage CSV file",
		Long: `Upload a usage data CSV file to Zuora (async).

The file is uploaded via multipart/form-data to POST /v1/usage.

Examples:
  zr usage post --file usage.csv`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.File == "" {
				return fmt.Errorf("--file is required")
			}
			return runPost(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "Path to CSV file to upload (required)")

	return cmd
}

func runPost(cmd *cobra.Command, opts *postOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(opts.File)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filepath.Base(opts.File))
	if err != nil {
		return fmt.Errorf("creating multipart form: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return fmt.Errorf("writing file data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("closing multipart writer: %w", err)
	}

	resp, err := client.Do("POST", "/v1/usage",
		api.WithBody(&buf),
		api.WithHeader("Content-Type", writer.FormDataContentType()),
		api.WithCheckSuccess(),
	)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Success", Value: getString(raw, "success")},
		{Key: "Check Import Status", Value: getString(raw, "checkImportStatus")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Usage file uploaded.\n")
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
