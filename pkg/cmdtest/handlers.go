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

// ObjectCRUDFailure returns a handler that responds 200 with the canonical
// Zuora Object-CRUD logical-failure envelope:
// {"Success":false,"Errors":[{"Code":code,"Message":message}]}. It is the
// uppercase counterpart to Reasons (which emits the v1 {"success":false} shape):
// Object-CRUD endpoints (/v1/object/...) report failures this way, so a test for
// such a command should model that shape, not the v1 one. code is a string
// because Object-CRUD error codes are string enums (e.g. "INVALID_VALUE").
func ObjectCRUDFailure(t *testing.T, code, message string) http.HandlerFunc {
	t.Helper()
	return OK(t, "", "", map[string]interface{}{
		"Success": false,
		"Errors":  []map[string]interface{}{{"Code": code, "Message": message}},
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

// Route returns a handler that dispatches by exact URL path to the matching
// handler in routes. It is for commands that call more than one endpoint (e.g.
// resolve an id via GET, then act via POST); single-endpoint tests should use
// OK/Status directly. A request to an unregistered path fails the test loudly —
// so a command hitting an unexpected endpoint is caught, not silently 404'd.
func Route(t *testing.T, routes map[string]http.HandlerFunc) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if h, ok := routes[r.URL.Path]; ok {
			h(w, r)
			return
		}
		assert.Failf(t, "unexpected request path",
			"no cmdtest.Route handler registered for %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}
}
