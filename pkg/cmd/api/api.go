// Package api implements the "zr api" raw API command.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/itchyny/gojq"
	internalapi "github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

type apiOptions struct {
	Factory  *factory.Factory
	Method   string
	Body     string
	Headers  []string
	Paginate bool
	JQ       string
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
			// Read --jq from global persistent flag
			opts.JQ, _ = cmd.Flags().GetString("jq")
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
	var output []byte
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
		output, err = json.MarshalIndent(allData, "", "  ")
		if err != nil {
			return err
		}
	} else {
		resp, err := client.Do(opts.Method, path, reqOpts...)
		if err != nil {
			return err
		}
		output = resp.Body
	}

	// Apply jq filter
	if opts.JQ != "" {
		filtered, err := filterJQ(output, opts.JQ)
		if err != nil {
			return err
		}
		output = filtered
	} else {
		// Pretty-print JSON
		var prettyJSON json.RawMessage
		if err := json.Unmarshal(output, &prettyJSON); err == nil {
			if pretty, err := json.MarshalIndent(prettyJSON, "", "  "); err == nil {
				output = pretty
			}
		}
	}

	fmt.Fprintln(f.IOStreams.Out, string(output))
	return nil
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

func filterJQ(data []byte, expr string) ([]byte, error) {
	query, err := gojq.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("parsing jq expression: %w", err)
	}

	var input interface{}
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("parsing JSON for jq: %w", err)
	}

	code, err := gojq.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("compiling jq expression: %w", err)
	}

	var results []string
	iter := code.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			// Handle gojq halt/halt_error: stop iteration cleanly
			var haltErr *gojq.HaltError
			if errors.As(err, &haltErr) {
				if haltErr.Value() != nil {
					return nil, fmt.Errorf("jq halt: %v", haltErr.Value())
				}
				break
			}
			return nil, fmt.Errorf("jq error: %w", err)
		}
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return nil, err
		}
		results = append(results, string(b))
	}

	return []byte(strings.Join(results, "\n")), nil
}
