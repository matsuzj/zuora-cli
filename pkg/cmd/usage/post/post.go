// Package post implements the "zr usage post" command.
package post

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
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

The file is uploaded via multipart/form-data to POST /v1/usage.`,
		Example: `  zr usage post --file usage.csv`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPost(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "Path to CSV file to upload (required)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func runPost(cmd *cobra.Command, opts *postOptions) error {
	f := opts.Factory
	file, err := os.Open(opts.File)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filepath.Base(opts.File))
	if err != nil {
		return fmt.Errorf("creating multipart form: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("writing file data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("closing multipart writer: %w", err)
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/usage",
		Body:   &buf,
		ReqOpts: []api.RequestOption{
			api.WithHeader("Content-Type", writer.FormDataContentType()),
		},
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
				{Key: "Check Import Status", Value: cmdutil.GetString(raw, "checkImportStatus")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return "Usage file uploaded.\n"
		},
	})
}
