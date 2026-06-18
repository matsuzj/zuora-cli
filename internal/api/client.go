package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/matsuzj/zuora-cli/internal/build"
)

// errRedirectRefused marks a redirect that CheckRedirect blocked (off-host or a
// cleartext downgrade). It is a deterministic policy rejection, so doWithRetry
// surfaces it immediately instead of retrying it like a transient transport error.
var errRedirectRefused = errors.New("redirect refused for credential safety")

// Client is an HTTP client for Zuora APIs.
type Client struct {
	baseURL       string
	httpClient    *http.Client
	tokenSource   func(context.Context) (string, error)
	refreshToken  func(context.Context) (string, error) // force refresh, bypassing cache
	zuoraVersion  string
	userAgent     string
	verbose       bool
	verboseBody   bool
	verboseWriter io.Writer
	readOnly      bool
	ctx           context.Context
	// sleep waits for d or until the context is cancelled; it is a seam so
	// tests can run the retry loop without real backoff delays.
	sleep func(ctx context.Context, d time.Duration) error
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithBaseURL sets the API base URL.
func WithBaseURL(u string) ClientOption {
	return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithTokenSource sets the token provider for authentication. The provider
// receives a context so a token fetch can be cancelled (e.g. Ctrl-C).
func WithTokenSource(fn func(context.Context) (string, error)) ClientOption {
	return func(c *Client) { c.tokenSource = fn }
}

// WithRefreshToken sets the force-refresh token provider (bypasses cache).
func WithRefreshToken(fn func(context.Context) (string, error)) ClientOption {
	return func(c *Client) { c.refreshToken = fn }
}

// WithZuoraVersion sets the Zuora-Version header.
func WithZuoraVersion(v string) ClientOption {
	return func(c *Client) { c.zuoraVersion = v }
}

// WithHTTPClient sets a custom http.Client. This is the deliberate test
// injection seam (httptest server clients, redirect-policy probes); the other
// runtime knobs go through the production Set* methods instead so tests
// exercise the same paths the CLI wires.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) { c.httpClient = hc }
}

// NewClient creates a new API client.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		userAgent:  "zuora-cli/" + build.Version,
		ctx:        context.Background(),
		sleep:      sleepWithContext,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.ctx == nil {
		c.ctx = context.Background()
	}
	if c.sleep == nil {
		c.sleep = sleepWithContext
	}
	// checkHost only validates the initial URL. Without a redirect policy, a 3xx
	// from the server (or a MITM) would have Go follow it and forward the request
	// body, Idempotency-Key, and Zuora-Entity-Ids off the configured host. Re-run
	// checkHost on every hop so redirects cannot carry the request off-host or
	// downgrade to cleartext. Guard WithHTTPClient callers too (unless they set
	// their own policy); copy the client first so a client shared across NewClient
	// instances isn't mutated and each gets a policy bound to ITS own baseURL.
	if c.httpClient != nil && c.httpClient.CheckRedirect == nil {
		cp := *c.httpClient
		cp.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("%w: stopped after 10 redirects", errRedirectRefused)
			}
			if err := c.checkHost(req.URL.String()); err != nil {
				// Wrap with the sentinel so doWithRetry fails fast instead of
				// retrying this deterministic policy rejection.
				return fmt.Errorf("%w: %v", errRedirectRefused, err)
			}
			return nil
		}
		c.httpClient = &cp
	}
	return c
}

// sleepWithContext waits for d or until ctx is cancelled, whichever comes first.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// newIdempotencyKey returns a random hex key for safely retrying mutations.
func newIdempotencyKey() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fall back to a time-based key; collisions are acceptable here because
		// the key only needs to be stable across retries of one request.
		return fmt.Sprintf("zr-%d", time.Now().UnixNano())
	}
	return "zr-" + hex.EncodeToString(b[:])
}

// SetZuoraVersion overrides the Zuora-Version header.
func (c *Client) SetZuoraVersion(v string) { c.zuoraVersion = v }

// SetVerbose enables verbose logging to the given writer.
func (c *Client) SetVerbose(w io.Writer) { c.verbose = true; c.verboseWriter = w }

// SetVerboseBody enables level-2 verbose: request/response bodies are logged
// (4KB cap, multipart skipped). Billing bodies are PII — this level is gated
// behind -vv / ZR_DEBUG=api and never enabled by plain --verbose.
func (c *Client) SetVerboseBody() { c.verboseBody = true }

// maxBodyLog caps how many body bytes level-2 verbose prints.
const maxBodyLog = 4096

// vlogBody prints a request (dir ">") or response (dir "<") body under
// level-2 verbose. Multipart payloads are skipped by Content-Type; larger
// bodies are truncated with an explicit marker.
func (c *Client) vlogBody(dir, contentType string, body []byte) {
	if !c.verboseBody || c.verboseWriter == nil || len(body) == 0 {
		return
	}
	// MIME media types are case-insensitive (e.g. "Multipart/Form-Data").
	if strings.HasPrefix(strings.ToLower(contentType), "multipart/") {
		fmt.Fprintf(c.verboseWriter, "%s [multipart body omitted]\n\n", dir)
		return
	}
	b := body
	truncated := false
	if len(b) > maxBodyLog {
		b = b[:maxBodyLog]
		truncated = true
	}
	fmt.Fprintf(c.verboseWriter, "%s %s\n", dir, string(b))
	if truncated {
		fmt.Fprintf(c.verboseWriter, "%s [body truncated at %d bytes]\n", dir, maxBodyLog)
	}
	fmt.Fprintln(c.verboseWriter)
}

// vlogf writes a gh-style diagnostic line (prefix "* ") to the verbose writer.
// No-op when verbose is off. Used by the retry loop to surface its decision
// points (backoff, Retry-After, token refresh) — values derived from secrets
// must never be passed here.
func (c *Client) vlogf(format string, args ...any) {
	if c.verbose && c.verboseWriter != nil {
		fmt.Fprintf(c.verboseWriter, "* "+format+"\n", args...)
	}
}

// SetReadOnly enables or disables read-only mode.
func (c *Client) SetReadOnly(v bool) { c.readOnly = v }

// SetContext sets the base context used for all requests.
func (c *Client) SetContext(ctx context.Context) {
	if ctx != nil {
		c.ctx = ctx
	}
}

// checkHost refuses a request whose absolute URL targets a host other than
// the configured base URL. Relative paths (resolved against baseURL) and
// same-host absolute URLs (e.g. a pagination nextPage) are allowed; a
// cross-host absolute URL is rejected so the bearer token is never sent off-host.
func (c *Client) checkHost(fullURL string) error {
	target, err := url.Parse(fullURL)
	if err != nil {
		return fmt.Errorf("invalid request URL: %w", err)
	}
	if target.Host == "" {
		return nil // relative — already rooted at the configured base URL
	}
	base, err := url.Parse(c.baseURL)
	if err != nil || base.Host == "" {
		return nil // no configured host to compare against
	}
	if !strings.EqualFold(target.Host, base.Host) {
		return fmt.Errorf("refusing to send credentials to %q: not the configured environment host %q", target.Host, base.Host)
	}
	// Refuse a cleartext downgrade: when the configured environment is https, a
	// same-host http target (e.g. `zr api http://...` or an http nextPage/redirect)
	// would put the bearer token on the wire in plaintext.
	if strings.EqualFold(base.Scheme, "https") && strings.EqualFold(target.Scheme, "http") {
		return fmt.Errorf("refusing to send credentials over cleartext http to %q (configured environment %q is https)", target.Host, base.Host)
	}
	return nil
}

// Do performs an HTTP request.
func (c *Client) Do(method, path string, opts ...RequestOption) (*Response, error) {
	if c.readOnly && !isReadOnlyAllowed(method, path) {
		return nil, &ReadOnlyError{Method: method, Path: path}
	}

	rc := newRequestConfig(opts)

	fullURL := c.buildURL(path, rc.query)

	// Never send credentials to a host other than the configured environment
	// host (e.g. a stray `zr api https://attacker/...`), which would leak the
	// bearer token off-host.
	if err := c.checkHost(fullURL); err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if rc.body != nil {
		bodyReader = rc.body
	}

	ctx := c.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Auth header
	if c.tokenSource != nil {
		token, err := c.tokenSource(ctx)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Standard headers
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if c.zuoraVersion != "" {
		req.Header.Set("Zuora-Version", c.zuoraVersion)
	}
	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Idempotency-Key for mutating methods so any retry (429/401/5xx) is
	// deduplicated server-side, preventing duplicate orders/payments/refunds.
	// carriesIdempotencyKey is the single source of truth for WHICH methods —
	// the SafeToRetry promise in retry.go depends on this being the same set.
	if carriesIdempotencyKey(method) {
		if req.Header.Get("Idempotency-Key") == "" {
			req.Header.Set("Idempotency-Key", newIdempotencyKey())
		}
	}

	// Custom headers (caller may override the above, e.g. multipart Content-Type)
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

	// Buffer the request body once so a transient HTTP-200 success=false retry
	// (below) can resend it: doWithRetry consumes req.Body, so without a fresh
	// reader per attempt a resend would transmit an EMPTY body for POST/PATCH.
	var reqBody []byte
	if req.Body != nil {
		reqBody, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("buffering request body: %w", err)
		}
	}

	// Shared retry budget for this request: doWithRetry (transport/429/401/5xx)
	// and this outer loop (HTTP-200 success=false transient codes) both draw
	// from it, so a request that hits both classes still makes at most
	// maxRetries+1 total requests rather than nesting two full retry loops.
	// doWithRetry decrements it per HTTP send; the success-envelope retry below
	// only continues while budget remains.
	budget := maxRetries + 1
	for attempt := 0; ; attempt++ {
		if attempt > 0 {
			total := backoffDuration(attempt)
			c.vlogf("retrying after %.1fs backoff (HTTP 200 success=false transient error)", total.Seconds())
			if err := c.sleep(ctx, total); err != nil {
				return nil, err
			}
		}
		if reqBody != nil {
			req.Body = io.NopCloser(bytes.NewReader(reqBody))
		}

		resp, err := c.doWithRetry(req, &budget)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
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
			c.vlogBody("<", resp.Header.Get("Content-Type"), body)
		}

		if resp.StatusCode >= 400 {
			return nil, parseAPIError(resp.StatusCode, body)
		}

		if rc.checkSuccess {
			if envErr := successEnvelopeError(resp.StatusCode, body); envErr != nil {
				if budget > 0 && isRetriableSuccessEnvelope(envErr, method) {
					continue
				}
				return nil, envErr
			}
		}

		return &Response{
			StatusCode: resp.StatusCode,
			Header:     resp.Header,
			Body:       body,
		}, nil
	}
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
