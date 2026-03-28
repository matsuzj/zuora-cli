package testutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

type registryTransport struct {
	base http.RoundTripper
}

func (t *registryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	handler := lookupHandler(req.URL.Host)
	if handler == nil {
		return t.base.RoundTrip(req)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder.Result(), nil
}

var (
	transportOnce sync.Once
	transportMu   sync.RWMutex
	transportMap  = map[string]http.Handler{}
	serverSeq     uint64
)

func installTransport() {
	transportOnce.Do(func() {
		base, _ := http.DefaultTransport.(*http.Transport)
		if base == nil {
			base = http.DefaultTransport.(*http.Transport)
		}
		http.DefaultTransport = &registryTransport{base: base.Clone()}
	})
}

func registerHandler(host string, handler http.Handler) {
	transportMu.Lock()
	defer transportMu.Unlock()
	transportMap[host] = handler
}

func unregisterHandler(host string) {
	transportMu.Lock()
	defer transportMu.Unlock()
	delete(transportMap, host)
}

func lookupHandler(host string) http.Handler {
	transportMu.RLock()
	defer transportMu.RUnlock()
	return transportMap[host]
}

// Server is a lightweight, no-network HTTP test server.
type Server struct {
	URL    string
	host   string
	client *http.Client
}

// Client returns an HTTP client configured for this server.
func (s *Server) Client() *http.Client {
	return s.client
}

// Close unregisters the server handler.
func (s *Server) Close() {
	unregisterHandler(s.host)
}

// NewServer creates a no-network HTTP test server backed by an in-process handler.
func NewServer(t *testing.T, handler http.Handler) *Server {
	t.Helper()

	installTransport()

	id := atomic.AddUint64(&serverSeq, 1)
	host := fmt.Sprintf("testserver-%d.invalid", id)
	registerHandler(host, handler)

	ts := &Server{
		URL:  "http://" + host,
		host: host,
		client: &http.Client{
			Transport: http.DefaultTransport,
		},
	}
	t.Cleanup(ts.Close)

	return ts
}

// NewJSONResponse builds an HTTP response with a JSON body for custom transports.
func NewJSONResponse(statusCode int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}
