// Package cmdutil provides shared utilities for CLI commands.
package cmdutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// ResolveBody resolves a body flag value to an io.Reader.
// Supported formats:
//   - "-" reads from stdin
//   - "@file" reads from the specified file
//   - any other string is treated as literal JSON
func ResolveBody(body string, stdin io.Reader) (io.Reader, error) {
	if body == "-" {
		return stdin, nil
	}
	if strings.HasPrefix(body, "@") {
		filePath := body[1:]
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading body file: %w", err)
		}
		return bytes.NewReader(data), nil
	}
	return strings.NewReader(body), nil
}
