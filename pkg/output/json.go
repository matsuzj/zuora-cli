package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
)

// PrintJSON writes pretty-printed JSON, optionally filtered by a jq expression.
func PrintJSON(ios *iostreams.IOStreams, data []byte, jqExpr string) error {
	if jqExpr != "" {
		return printJQ(ios, data, jqExpr)
	}
	// Pretty-print
	var v json.RawMessage
	if err := json.Unmarshal(data, &v); err != nil {
		// Not valid JSON, write as-is
		fmt.Fprintln(ios.Out, string(data))
		return nil
	}
	pretty, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintln(ios.Out, string(data))
		return nil
	}
	fmt.Fprintln(ios.Out, string(pretty))
	return nil
}

func printJQ(ios *iostreams.IOStreams, data []byte, expr string) error {
	query, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Errorf("parsing jq expression: %w", err)
	}

	var input interface{}
	if err := json.Unmarshal(data, &input); err != nil {
		return fmt.Errorf("parsing JSON for jq: %w", err)
	}

	code, err := gojq.Compile(query)
	if err != nil {
		return fmt.Errorf("compiling jq expression: %w", err)
	}

	var results []string
	iter := code.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			var haltErr *gojq.HaltError
			if errors.As(err, &haltErr) {
				if haltErr.Value() != nil {
					return fmt.Errorf("jq halt: %v", haltErr.Value())
				}
				break
			}
			return fmt.Errorf("jq error: %w", err)
		}
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		results = append(results, string(b))
	}

	fmt.Fprintln(ios.Out, strings.Join(results, "\n"))
	return nil
}
