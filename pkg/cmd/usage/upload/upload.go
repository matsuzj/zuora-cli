// Package upload implements the "zr usage upload" command.
package upload

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

type uploadOptions struct {
	Factory *factory.Factory
	File    string
}

// NewCmdUpload creates the usage upload command. It was renamed from
// `usage post` — "post" collided with the document-lifecycle verb used by
// `invoice post` / `billrun post` (transition to Posted), which is unrelated to
// uploading a usage file. The old `post` name stays as a deprecated alias.
func NewCmdUpload(f *factory.Factory) *cobra.Command {
	opts := &uploadOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "upload",
		Aliases: []string{"post"},
		Short:   "Upload a usage CSV file",
		Long: `Upload a usage data CSV file to Zuora (async).

The file is uploaded via multipart/form-data to POST /v1/usage.`,
		Example: `  zr usage upload --file usage.csv`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Deprecation nudge only when invoked by the old name; `upload`
			// itself stays quiet. (cobra's Deprecated field would warn for both.)
			if cmd.CalledAs() == "post" {
				fmt.Fprintln(f.IOStreams.ErrOut, "warning: 'usage post' is deprecated; use 'usage upload'.")
			}
			return runUpload(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "Path to CSV file to upload (required)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func runUpload(cmd *cobra.Command, opts *uploadOptions) error {
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
