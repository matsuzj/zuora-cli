package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// exitCodeOf mirrors cmd/zr/main.go's exitCode(): an error implementing
// ExitCode() returns its code, otherwise a general error is exit 1.
func exitCodeOf(err error) int {
	var ec interface{ ExitCode() int }
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return 1
}

func TestAPIError_TransportFailure_ExitCode(t *testing.T) {
	assert.Equal(t, 1, (&APIError{StatusCode: 0}).ExitCode(),
		"a transport failure (no HTTP response) is not a 4xx client error; it must exit 1")
}

// TestClient_TransportError_ExitCodeConsistentAcrossMethods covers the audit
// finding that the same transport failure exited 1 on idempotent GET/DELETE but
// 3 (mislabeled "4xx client error") on POST/PUT/PATCH. Both must now yield 1.
func TestClient_TransportError_ExitCodeConsistentAcrossMethods(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	addr := srv.URL
	srv.Close() // closing guarantees connection-refused at addr

	client := newNoSleepClient(WithBaseURL(addr))

	_, postErr := client.Post("/v1/x", strings.NewReader(`{}`))
	require.Error(t, postErr)
	assert.Equal(t, 1, exitCodeOf(postErr), "POST transport failure must exit 1")

	_, getErr := client.Get("/v1/x")
	require.Error(t, getErr)
	assert.Equal(t, 1, exitCodeOf(getErr), "GET transport failure must exit 1 (consistent with POST)")
}
