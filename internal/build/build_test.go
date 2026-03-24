package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultVersion(t *testing.T) {
	assert.Equal(t, "dev", Version)
}

func TestDefaultCommitAndDate(t *testing.T) {
	assert.Empty(t, Commit)
	assert.Empty(t, Date)
}
