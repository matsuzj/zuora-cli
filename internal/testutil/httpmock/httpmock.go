package httpmock

import (
	"fmt"
	"net/http"
	stdhttptest "net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	serverSeq uint64
	mu        sync.RWMutex
	clients   = map[string]*http.Client{}
)

// Server is a lightweight httptest-compatible server backed by an in-memory transport.
type Server struct {
	URL    string
	client *http.Client
}

// NewServer creates a new in-memory test server.
func NewServer(handler http.Handler) *Server {
	id := atomic.AddUint64(&serverSeq, 1)
	baseURL := fmt.Sprintf("http://mock-%d.test", id)
	client := &http.Client{
		Transport: roundTripper{handler: handler},
	}

	mu.Lock()
	clients[baseURL] = client
	mu.Unlock()

	return &Server{
		URL:    baseURL,
		client: client,
	}
}

// Client returns an HTTP client that routes requests to the in-memory handler.
func (s *Server) Client() *http.Client {
	return s.client
}

// Close unregisters the server.
func (s *Server) Close() {
	mu.Lock()
	delete(clients, strings.TrimRight(s.URL, "/"))
	mu.Unlock()
}

// ClientForURL returns an in-memory client for a registered test server URL.
func ClientForURL(rawURL string) (*http.Client, bool) {
	mu.RLock()
	client, ok := clients[strings.TrimRight(rawURL, "/")]
	mu.RUnlock()
	return client, ok
}

type roundTripper struct {
	handler http.Handler
}

func (rt roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	if req.Body != nil {
		req2.Body = req.Body
	}
	req2.URL = cloneURL(req.URL)
	req2.RequestURI = req2.URL.RequestURI()

	rec := stdhttptest.NewRecorder()
	rt.handler.ServeHTTP(rec, req2)

	resp := rec.Result()
	resp.Request = req
	return resp, nil
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	copied := *u
	return &copied
}
