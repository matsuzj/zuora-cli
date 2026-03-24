// Package factory provides dependency injection for CLI commands.
package factory

import (
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
)

// Factory provides shared dependencies to all commands.
type Factory struct {
	IOStreams *iostreams.IOStreams
}

// New creates a Factory with real (system) dependencies.
func New() *Factory {
	return &Factory{
		IOStreams: iostreams.System(),
	}
}
