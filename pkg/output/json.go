package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
)

// emptyBody reports whether data has no content (e.g. an HTTP 204 or empty
// 200) — every output path treats that as "nothing to print", success, so the
// exit code stays consistent across modes.
func emptyBody(data []byte) bool {
	return len(bytes.TrimSpace(data)) == 0
}

// prettyJSON validates data as a single JSON value and returns it re-indented.
// Callers keep their own fallback behavior on error (raw passthrough for the
// api escape hatch, stderr+non-zero for strict JSON output).
func prettyJSON(data []byte) (string, error) {
	var v json.RawMessage
	if err := json.Unmarshal(data, &v); err != nil {
		return "", err
	}
	pretty, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(pretty), nil
}

// decodeJSONPreservingNumbers decodes JSON into a generic value using
// json.Number for all numbers, so large integers (e.g. Zuora's 19-digit IDs)
// and high-precision amounts are not silently rounded by float64. gojq accepts
// json.Number natively and promotes oversized integers to *big.Int.
func decodeJSONPreservingNumbers(data []byte) (interface{}, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v interface{}
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	// Reject trailing garbage after the first value, so "{...} junk" is not
	// silently accepted as valid (matches strict json.Unmarshal behavior).
	if dec.More() {
		return nil, fmt.Errorf("unexpected trailing data after JSON value")
	}
	return v, nil
}

// PrintRawOrJSON pretty-prints data when it is valid JSON, otherwise writes the
// raw body to stdout unchanged and succeeds. It is for the raw `api` escape
// hatch, whose response may legitimately be non-JSON (text, CSV, a proxy's HTML
// success page), where failing would lose the body and break `zr api ... > file`.
func PrintRawOrJSON(ios *iostreams.IOStreams, data []byte) error {
	if emptyBody(data) {
		return nil
	}
	pretty, err := prettyJSON(data)
	if err != nil {
		// Not JSON — pass the body through verbatim (no added newline), so a
		// binary or exact-byte body redirected from `zr api` is not corrupted.
		_, werr := ios.Out.Write(data)
		return werr
	}
	fmt.Fprintln(ios.Out, pretty)
	return nil
}

// PrintJSON writes pretty-printed JSON, optionally filtered by a jq expression.
func PrintJSON(ios *iostreams.IOStreams, data []byte, jqExpr string) error {
	if jqExpr != "" {
		return printJQ(ios, data, jqExpr)
	}
	// Pretty-print. An empty body (e.g. HTTP 204) is not an error.
	if emptyBody(data) {
		return nil
	}
	pretty, err := prettyJSON(data)
	if err != nil {
		// Not valid JSON: echo the raw body to stderr and fail, so scripts and
		// downstream JSON consumers can detect it via a non-zero exit code
		// instead of receiving a corrupt stream on stdout.
		fmt.Fprintln(ios.ErrOut, string(data))
		return fmt.Errorf("response is not valid JSON")
	}
	fmt.Fprintln(ios.Out, pretty)
	return nil
}

func printJQ(ios *iostreams.IOStreams, data []byte, expr string) error {
	query, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Errorf("parsing jq expression: %w", err)
	}

	// An empty body (e.g. HTTP 204) has nothing to filter — succeed silently,
	// matching the pretty-print (PrintJSON) and raw (PrintRawOrJSON) paths so the
	// exit code is consistent across output modes.
	if emptyBody(data) {
		return nil
	}

	input, err := decodeJSONPreservingNumbers(data)
	if err != nil {
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

	if len(results) > 0 {
		fmt.Fprintln(ios.Out, strings.Join(results, "\n"))
	}
	return nil
}
