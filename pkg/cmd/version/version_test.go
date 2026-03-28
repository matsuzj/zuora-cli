package version

import (
	"testing"

	"github.com/matsuzj/zuora-cli/internal/build"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestVersionOutput(t *testing.T) {
	originalVersion := build.Version
	originalCommit := build.Commit
	originalDate := build.Date
	t.Cleanup(func() {
		build.Version = originalVersion
		build.Commit = originalCommit
		build.Date = originalDate
	})

	build.Version = "1.2.3"
	build.Commit = "abc123"
	build.Date = "2026-03-29T00:00:00Z"

	ios, _, out, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	cmd := NewCmdVersion(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Equal(t, "zr version 1.2.3 (commit: abc123) (built: 2026-03-29T00:00:00Z)\n", out.String())
}
