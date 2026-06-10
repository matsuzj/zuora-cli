package cmdtest

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// OK returns a handler that optionally asserts the HTTP method and URL path
// ("" skips either assertion), then responds 200 with body JSON-encoded.
// It captures t at construction time because the handler runs on the test
// server's goroutine.
func OK(t *testing.T, method, path string, body interface{}) http.HandlerFunc {
	t.Helper()
	return Status(t, method, path, http.StatusOK, body)
}

// Reasons returns a handler that responds 200 with the canonical Zuora v1
// logical-failure envelope: {"success":false,"reasons":[{code,message}]}.
// code is interface{} because Zuora emits both numeric and string codes.
func Reasons(t *testing.T, code interface{}, message string) http.HandlerFunc {
	t.Helper()
	return OK(t, "", "", map[string]interface{}{
		"success": false,
		"reasons": []map[string]interface{}{{"code": code, "message": message}},
	})
}

// Status is OK with an explicit status code, for 4xx/5xx and 204 shapes.
// body may be nil to send no body.
func Status(t *testing.T, method, path string, statusCode int, body interface{}) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if method != "" {
			assert.Equal(t, method, r.Method)
		}
		if path != "" {
			assert.Equal(t, path, r.URL.Path)
		}
		w.WriteHeader(statusCode)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}
}
