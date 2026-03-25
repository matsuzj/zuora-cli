// Package api implements the "zr api" raw API command.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	internalapi "github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type apiOptions struct {
	Factory  *factory.Factory
	Method   string
	Body     string
	Headers  []string
	Paginate bool
	JQ       string
	Template string
}

// NewCmdAPI creates the api command.
func NewCmdAPI(f *factory.Factory) *cobra.Command {
	opts := &apiOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "api <path>",
		Short: "Make a raw API request",
		Long: `Make an authenticated HTTP request to the Zuora API.

Examples:
  zr api /v1/accounts                          # GET (default)
  zr api -X POST /v1/orders --body @order.json # POST with file body
  zr api /v1/accounts --jq '.accounts[].name'  # jq filter
  zr api /v1/accounts --paginate               # Fetch all pages`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Method = strings.ToUpper(opts.Method)
			// Read --jq and --template from global persistent flags
			opts.JQ, _ = cmd.Flags().GetString("jq")
			opts.Template, _ = cmd.Flags().GetString("template")
			return runAPI(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.Method, "method", "X", "GET", "HTTP method")
	cmd.Flags().StringVarP(&opts.Body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")
	cmd.Flags().StringArrayVarP(&opts.Headers, "header", "H", nil, "Additional headers (key:value)")
	cmd.Flags().BoolVar(&opts.Paginate, "paginate", false, "Fetch all pages automatically")
	// NOTE: --jq is inherited from root persistent flags, not defined locally

	return cmd
}

func runAPI(opts *apiOptions, path string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	// Build request options
	var reqOpts []internalapi.RequestOption

	// Resolve body
	if opts.Body != "" {
		bodyReader, err := resolveBody(opts.Body, f.IOStreams.In)
		if err != nil {
			return err
		}
		reqOpts = append(reqOpts, internalapi.WithBody(bodyReader))
	}

	// Parse custom headers
	for _, h := range opts.Headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid header format %q (expected key:value)", h)
		}
		reqOpts = append(reqOpts, internalapi.WithHeader(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])))
	}

	// Execute request
	var result []byte
	if opts.Paginate {
		pages, err := client.DoPaginated(opts.Method, path, reqOpts...)
		if err != nil {
			return err
		}
		var allData []json.RawMessage
		for _, page := range pages {
			var items []json.RawMessage
			if err := json.Unmarshal(page, &items); err != nil {
				allData = append(allData, page)
			} else {
				allData = append(allData, items...)
			}
		}
		result, err = json.MarshalIndent(allData, "", "  ")
		if err != nil {
			return err
		}
	} else {
		resp, err := client.Do(opts.Method, path, reqOpts...)
		if err != nil {
			return err
		}
		result = resp.Body
	}

	// Output via shared formatter (precedence: --jq > --template > pretty-print)
	if opts.JQ != "" {
		return output.PrintJSON(f.IOStreams, result, opts.JQ)
	}
	if opts.Template != "" {
		return output.PrintTemplate(f.IOStreams, result, opts.Template)
	}
	return output.PrintJSON(f.IOStreams, result, "")
}

func resolveBody(body string, stdin io.Reader) (io.Reader, error) {
	if body == "-" {
		return stdin, nil
	}
	if strings.HasPrefix(body, "@") {
		filePath := body[1:]
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading body file: %w", err)
		}
		return strings.NewReader(string(data)), nil
	}
	return strings.NewReader(body), nil
}
