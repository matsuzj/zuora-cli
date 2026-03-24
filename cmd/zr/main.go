package main

import (
	"fmt"
	"os"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/root"
)

func main() {
	f := factory.New()
	rootCmd := root.NewCmdRoot(f)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(f.IOStreams.ErrOut, "Error: %s\n", err)
		os.Exit(1)
	}
}
