// Package cmdutil provides shared utilities for CLI commands.
package cmdutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// ResolveBody resolves a body flag value to an io.Reader and validates that the
// resolved content is well-formed JSON. Every --body command targets a Zuora
// JSON endpoint, so a malformed body is caught here with a clear local error
// instead of a confusing server-side 4xx (F-22).
//
// Supported formats:
//   - "-" reads from stdin
//   - "@file" reads from the specified file
//   - any other string is treated as literal JSON
func ResolveBody(body string, stdin io.Reader) (io.Reader, error) {
	// An explicitly-empty value (--body "", or --body "$VAR" with VAR unset)
	// satisfies cobra's required check (the flag WAS provided) but cannot be
	// a meaningful request body — fail fast locally instead of sending an
	// empty body to Zuora (Codex, P5-2).
	if body == "" {
		return nil, fmt.Errorf("request body is empty")
	}

	var data []byte
	switch {
	case body == "-":
		b, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("reading body from stdin: %w", err)
		}
		data = b
	case strings.HasPrefix(body, "@"):
		b, err := os.ReadFile(body[1:])
		if err != nil {
			return nil, fmt.Errorf("reading body file: %w", err)
		}
		data = b
	default:
		data = []byte(body)
	}

	if !json.Valid(data) {
		return nil, fmt.Errorf("request body is not valid JSON")
	}
	return bytes.NewReader(data), nil
}
