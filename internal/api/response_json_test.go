package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResponse_JSON_Unmarshals pins the happy path of the JSON helper: the body
// is decoded into the target struct.
func TestResponse_JSON_Unmarshals(t *testing.T) {
	r := &Response{Body: []byte(`{"id":"A-001","active":true}`)}
	var out struct {
		ID     string `json:"id"`
		Active bool   `json:"active"`
	}
	require.NoError(t, r.JSON(&out))
	assert.Equal(t, "A-001", out.ID)
	assert.True(t, out.Active)
}

// TestResponse_JSON_DecodeError pins the error branch: invalid JSON in the body
// surfaces as an error rather than being silently ignored.
func TestResponse_JSON_DecodeError(t *testing.T) {
	r := &Response{Body: []byte(`{not valid json`)}
	var out map[string]any
	require.Error(t, r.JSON(&out))
}
