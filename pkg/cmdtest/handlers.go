package cmdtest

import (
	"encoding/json"
	"io"
	"net/http"
	"sync/atomic"
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

// Sequence returns a handler that serves request N with handlers[N]; once the
// sequence is exhausted the LAST handler repeats, so a polling command can
// keep re-reading the terminal state. The counter is atomic because handlers
// run on the test server's goroutines. Compose with Route for multi-endpoint
// polling flows (submit endpoint -> OK, status endpoint -> Sequence(inProgress,
// done)). At least one handler is required.
func Sequence(handlers ...http.HandlerFunc) http.HandlerFunc {
	var n atomic.Int64
	return func(w http.ResponseWriter, r *http.Request) {
		i := int(n.Add(1)) - 1
		if i >= len(handlers) {
			i = len(handlers) - 1
		}
		handlers[i](w, r)
	}
}

// ZOQLPages returns a handler for the canonical two-page ZOQL pagination
// contract: call 1 must be POST /v1/action/query and gets page1 records with
// done:false plus a queryLocator; call 2 must be POST /v1/action/queryMore
// echoing that locator and gets page2 records with done:true. Any further
// request fails the test, and a t.Cleanup fails it if fewer than two calls
// arrived — so a caller keeps the "exactly two requests" guarantee without a
// hand-rolled counter. (This models the action/query family only; the
// nextPage-URL style of listcmd/`zr api --paginate` has a different envelope —
// build those from Sequence instead.)
func ZOQLPages(t *testing.T, page1, page2 []map[string]interface{}) http.HandlerFunc {
	t.Helper()
	const locator = "cmdtest-locator-1"
	var n atomic.Int64
	t.Cleanup(func() {
		assert.EqualValues(t, 2, n.Load(), "ZOQL pagination must fetch exactly two pages")
	})
	return func(w http.ResponseWriter, r *http.Request) {
		switch n.Add(1) {
		case 1:
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/action/query", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"records": page1, "size": len(page1), "done": false, "queryLocator": locator,
			})
		case 2:
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/action/queryMore", r.URL.Path)
			var body map[string]interface{}
			if assert.NoError(t, json.NewDecoder(r.Body).Decode(&body), "decoding queryMore body") {
				assert.Equal(t, locator, body["queryLocator"], "queryMore must echo the page-1 locator")
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"records": page2, "size": len(page2), "done": true,
			})
		default:
			assert.Failf(t, "unexpected request after the final ZOQL page",
				"%s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// Expect is a declarative request matcher + responder for the assertions OK and
// Status don't cover: request body, headers, and query. A zero field is not
// asserted (Expect{Path: "/v1/orders"} checks only the path), so it scales from
// "just the path" up to a full request contract without a hand-rolled handler.
type Expect struct {
	Method   string            // asserted when non-empty
	Path     string            // asserted when non-empty
	Query    map[string]string // each key asserted when the map is non-nil
	Headers  map[string]string // each header asserted when the map is non-nil
	JSONBody string            // request body asserted JSON-equal when non-empty
	Status   int               // response status (0 -> 200)
	Respond  interface{}       // response body, JSON-encoded (nil -> no body)
}

// Handler builds the http.HandlerFunc for Run from e, and arms a t.Cleanup that
// fails the test if the handler is never reached — so a command that short-
// circuits before its HTTP call (yet was given a request-asserting handler) is
// caught instead of passing on assertions that never ran. Assertions use assert
// (not require) because the handler runs on the test server's goroutine.
func (e Expect) Handler(t *testing.T) http.HandlerFunc {
	t.Helper()
	var reached atomic.Bool
	t.Cleanup(func() {
		assert.True(t, reached.Load(),
			"expected request never arrived — the command made no matching HTTP call")
	})
	return func(w http.ResponseWriter, r *http.Request) {
		reached.Store(true)
		if e.Method != "" {
			assert.Equal(t, e.Method, r.Method)
		}
		if e.Path != "" {
			assert.Equal(t, e.Path, r.URL.Path)
		}
		for k, v := range e.Query {
			assert.Equal(t, v, r.URL.Query().Get(k), "query param %q", k)
		}
		for k, v := range e.Headers {
			assert.Equal(t, v, r.Header.Get(k), "header %q", k)
		}
		if e.JSONBody != "" {
			body, err := io.ReadAll(r.Body)
			if assert.NoError(t, err, "reading request body") {
				assert.JSONEq(t, e.JSONBody, string(body))
			}
		}
		status := e.Status
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		if e.Respond != nil {
			_ = json.NewEncoder(w).Encode(e.Respond)
		}
	}
}
