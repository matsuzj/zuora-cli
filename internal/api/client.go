package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/matsuzj/zuora-cli/internal/testutil/httpmock"
)

// Client is an HTTP client for Zuora APIs.
type Client struct {
	baseURL       string
	httpClient    *http.Client
	httpClientSet bool
	tokenSource   func() (string, error)
	refreshToken  func() (string, error) // force refresh, bypassing cache
	zuoraVersion  string
	verbose       bool
	verboseWriter io.Writer
	readOnly      bool
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithBaseURL sets the API base URL.
func WithBaseURL(u string) ClientOption {
	return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithTokenSource sets the token provider for authentication.
func WithTokenSource(fn func() (string, error)) ClientOption {
	return func(c *Client) { c.tokenSource = fn }
}

// WithRefreshToken sets the force-refresh token provider (bypasses cache).
func WithRefreshToken(fn func() (string, error)) ClientOption {
	return func(c *Client) { c.refreshToken = fn }
}

// WithZuoraVersion sets the Zuora-Version header.
func WithZuoraVersion(v string) ClientOption {
	return func(c *Client) { c.zuoraVersion = v }
}

// WithVerbose enables verbose request/response logging.
func WithVerbose(w io.Writer) ClientOption {
	return func(c *Client) { c.verbose = true; c.verboseWriter = w }
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
		c.httpClientSet = true
	}
}

// WithReadOnly enables read-only mode, blocking write operations.
func WithReadOnly() ClientOption {
	return func(c *Client) { c.readOnly = true }
}

// NewClient creates a new API client.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	if !c.httpClientSet {
		if hc, ok := httpmock.ClientForURL(c.baseURL); ok {
			c.httpClient = hc
		}
	}
	return c
}

// SetZuoraVersion overrides the Zuora-Version header.
func (c *Client) SetZuoraVersion(v string) { c.zuoraVersion = v }

// SetVerbose enables verbose logging to the given writer.
func (c *Client) SetVerbose(w io.Writer) { c.verbose = true; c.verboseWriter = w }

// SetReadOnly enables or disables read-only mode.
func (c *Client) SetReadOnly(v bool) { c.readOnly = v }

// Do performs an HTTP request.
func (c *Client) Do(method, path string, opts ...RequestOption) (*Response, error) {
	if c.readOnly && !isReadOnlyAllowed(method, path) {
		return nil, &ReadOnlyError{Method: method, Path: path}
	}

	rc := newRequestConfig(opts)

	fullURL := c.buildURL(path, rc.query)

	var bodyReader io.Reader
	if rc.body != nil {
		bodyReader = rc.body
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Auth header
	if c.tokenSource != nil {
		token, err := c.tokenSource()
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Standard headers
	if c.zuoraVersion != "" {
		req.Header.Set("Zuora-Version", c.zuoraVersion)
	}
	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Custom headers
	for k, v := range rc.headers {
		req.Header.Set(k, v)
	}

	if c.verbose && c.verboseWriter != nil {
		fmt.Fprintf(c.verboseWriter, "> %s %s\n", method, fullURL)
		for k, vs := range req.Header {
			for _, v := range vs {
				if k == "Authorization" {
					fmt.Fprintf(c.verboseWriter, "> %s: Bearer ***\n", k)
				} else {
					fmt.Fprintf(c.verboseWriter, "> %s: %s\n", k, v)
				}
			}
		}
		fmt.Fprintln(c.verboseWriter)
	}

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if c.verbose && c.verboseWriter != nil {
		fmt.Fprintf(c.verboseWriter, "< HTTP %d\n", resp.StatusCode)
		for k, vs := range resp.Header {
			for _, v := range vs {
				fmt.Fprintf(c.verboseWriter, "< %s: %s\n", k, v)
			}
		}
		fmt.Fprintln(c.verboseWriter)
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, body)
	}

	// Zuora success flag check: some endpoints return HTTP 200 with {"success": false}
	// Note: v1 REST API uses lowercase "success", Object CRUD API uses uppercase "Success"
	if rc.checkSuccess {
		var envelope struct {
			Success      *bool `json:"success"`
			SuccessUpper *bool `json:"Success"`
		}
		if json.Unmarshal(body, &envelope) == nil {
			if envelope.Success != nil && !*envelope.Success {
				return nil, parseAPIError(resp.StatusCode, body)
			}
			if envelope.SuccessUpper != nil && !*envelope.SuccessUpper {
				return nil, parseAPIError(resp.StatusCode, body)
			}
		}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       body,
	}, nil
}

// Get performs a GET request.
func (c *Client) Get(path string, opts ...RequestOption) (*Response, error) {
	return c.Do(http.MethodGet, path, opts...)
}

// Post performs a POST request.
func (c *Client) Post(path string, body io.Reader, opts ...RequestOption) (*Response, error) {
	return c.Do(http.MethodPost, path, append(opts, WithBody(body))...)
}

// Put performs a PUT request.
func (c *Client) Put(path string, body io.Reader, opts ...RequestOption) (*Response, error) {
	return c.Do(http.MethodPut, path, append(opts, WithBody(body))...)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string, opts ...RequestOption) (*Response, error) {
	return c.Do(http.MethodDelete, path, opts...)
}

// Patch performs a PATCH request.
func (c *Client) Patch(path string, body io.Reader, opts ...RequestOption) (*Response, error) {
	return c.Do(http.MethodPatch, path, append(opts, WithBody(body))...)
}

func (c *Client) buildURL(path string, query url.Values) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		// Absolute URL (e.g., nextPage) — already contains all query params.
		// Do not merge request-level query params to avoid duplicates on pagination.
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// Parse the path to merge any existing query params with new ones
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return c.baseURL + path
	}
	if len(query) > 0 {
		q := u.Query()
		for k, vs := range query {
			if _, exists := q[k]; exists {
				// Key already present in the URL (e.g., from a nextPage path).
				// Skip to avoid duplicating query params on paginated requests.
				continue
			}
			for _, v := range vs {
				q.Add(k, v)
			}
		}
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// readOnlyPOSTAllowList contains POST endpoints that are read-only (exact match).
var readOnlyPOSTAllowList = []string{
	// ZOQL query
	"v1/action/query",
	"v1/action/querymore",
	// Commerce API query/list (POST but read-only)
	"commerce/charges/query",
	"commerce/plans/query",
	"commerce/plans/list",
	"commerce/purchase-options/list",
	"commerce/legacy/products/list",
	// Preview (no data mutation, simulation only)
	"v1/orders/preview",
	"v1/async/orders/preview",
	"v1/subscriptions/preview",
}

// readOnlyPOSTPatterns contains POST endpoints with dynamic path segments (regex match).
var readOnlyPOSTPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^v1/subscriptions/[^/]+/preview$`), // preview-change
	regexp.MustCompile(`^meters/[^/]+/summary$`),           // meter summary (read-only)
}

// extractPath normalises a request path for allowlist matching.
// It handles absolute URLs, strips query parameters, removes leading slashes,
// and lowercases the result.
func extractPath(rawPath string) string {
	p := rawPath
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		if u, err := url.Parse(p); err == nil {
			p = u.Path
		}
	}
	if idx := strings.Index(p, "?"); idx >= 0 {
		p = p[:idx]
	}
	return strings.ToLower(strings.TrimLeft(p, "/"))
}

// isReadOnlyAllowed returns true if the given method+path combination is allowed
// in read-only mode. GET/HEAD/OPTIONS are always allowed. POST is allowed only
// for allowlisted read-only endpoints. PUT/DELETE/PATCH are always blocked.
func isReadOnlyAllowed(method, path string) bool {
	m := strings.ToUpper(method)
	switch m {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	case http.MethodPost:
		p := extractPath(path)
		for _, allowed := range readOnlyPOSTAllowList {
			if p == allowed {
				return true
			}
		}
		for _, re := range readOnlyPOSTPatterns {
			if re.MatchString(p) {
				return true
			}
		}
		return false
	default:
		// PUT, DELETE, PATCH, etc.
		return false
	}
}
