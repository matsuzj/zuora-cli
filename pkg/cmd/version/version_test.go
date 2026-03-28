package version

import (
	"testing"

	"github.com/matsuzj/zuora-cli/internal/build"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionOutput(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	cmd := NewCmdVersion(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Contains(t, out.String(), "zr version")
}

func TestRunVersion_IncludesBuildMetadata(t *testing.T) {
	origVersion := build.Version
	origCommit := build.Commit
	origDate := build.Date
	t.Cleanup(func() {
		build.Version = origVersion
		build.Commit = origCommit
		build.Date = origDate
	})

	build.Version = "1.2.3"
	build.Commit = "abc123"
	build.Date = "2026-03-28T00:00:00Z"

	ios, _, out, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	require.NoError(t, runVersion(f))
	assert.Equal(t, "zr version 1.2.3 (commit: abc123) (built: 2026-03-28T00:00:00Z)\n", out.String())
}
