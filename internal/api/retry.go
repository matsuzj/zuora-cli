package api

import (
	"bytes"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	maxRetries = 3
	// maxRetryAfter caps a server-supplied Retry-After so a hostile or
	// misconfigured value (e.g. 86400) cannot hang the CLI indefinitely.
	maxRetryAfter = 60 * time.Second
)

// isIdempotent returns true for HTTP methods that are safe to retry on 5xx.
// PUT is retried but does NOT carry an Idempotency-Key (Zuora rejects PUT+key,
// HTTP 400). Action-style PUTs (e.g. invoice write-off) therefore keep a small
// double-apply risk if a successful call's response is lost — a known tradeoff.
func isIdempotent(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions || method == http.MethodPut || method == http.MethodDelete
}

func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	// Save body for replay on retry (POST/PUT/PATCH)
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	var lastErr error
	tokenRefreshed := false
	skipBackoff := false

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Honor cancellation before each attempt. Return the cancellation error
		// (not a stale prior API error) so Ctrl-C is classified as cancellation.
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if attempt > 0 {
			if !skipBackoff {
				backoff := time.Duration(1<<(attempt-1)) * time.Second
				jitter := time.Duration(rand.Int64N(int64(backoff / 2)))
				if err := c.sleep(ctx, backoff+jitter); err != nil {
					return nil, err
				}
			}
			skipBackoff = false

			if bodyBytes != nil {
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Only retry transport errors for idempotent methods. For POST/PATCH
			// the request may have reached the server, so we do not auto-retry;
			// the caller can safely re-run (the request carries an
			// Idempotency-Key) — surface that guidance.
			if !isIdempotent(req.Method) {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return nil, ctxErr
				}
				return nil, &APIError{Message: err.Error(), SafeToRetry: true}
			}
			lastErr = err
			continue
		}

		switch {
		case resp.StatusCode == http.StatusTooManyRequests:
			// 429 means the request was rate-limited, not processed. Mutations
			// carry an Idempotency-Key (set in Do) so a replay is safe.
			lastErr = readAPIError(resp)
			if d, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
				if d > maxRetryAfter {
					d = maxRetryAfter
				}
				if err := c.sleep(ctx, d); err != nil {
					// Cancelled (Ctrl-C) while honoring Retry-After: surface the
					// cancellation, not a stale "HTTP 429" — matching the backoff
					// path above and the loop-top context check.
					return nil, err
				}
				skipBackoff = true
			}
			continue

		case resp.StatusCode == http.StatusUnauthorized && !tokenRefreshed:
			// 401 happens before processing (auth is checked first), so a
			// resend after refresh is safe for all methods.
			lastErr = readAPIError(resp)
			refreshFn := c.refreshToken
			if refreshFn == nil {
				refreshFn = c.tokenSource
			}
			if refreshFn != nil {
				token, err := refreshFn(ctx)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Authorization", "Bearer "+token)
				tokenRefreshed = true
				skipBackoff = true
				continue
			}
			return nil, lastErr

		case resp.StatusCode >= 500 && isIdempotent(req.Method):
			// Only retry 5xx for idempotent methods to avoid duplicate mutations.
			lastErr = readAPIError(resp)
			continue

		case resp.StatusCode >= 500:
			// Non-idempotent 5xx: do not auto-retry (the mutation may have been
			// applied), but mark it safe to re-run thanks to the Idempotency-Key.
			apiErr := readAPIError(resp)
			if ae, ok := apiErr.(*APIError); ok {
				ae.SafeToRetry = true
			}
			return nil, apiErr

		default:
			return resp, nil
		}
	}

	return nil, lastErr
}

// readAPIError reads and closes the response body, returning a parsed APIError
// that preserves Zuora's real error code/message instead of a generic string.
func readAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return parseAPIError(resp.StatusCode, body)
}

// parseRetryAfter parses a Retry-After header, which may be either a number of
// seconds ("120") or an HTTP date ("Wed, 21 Oct 2026 07:28:00 GMT"). It returns
// the wait duration and whether a usable value was found. Negative/past values
// clamp to zero.
func parseRetryAfter(v string) (time.Duration, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(v); err == nil {
		if secs < 0 {
			secs = 0
		}
		return time.Duration(secs) * time.Second, true
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d < 0 {
			d = 0
		}
		return d, true
	}
	return 0, false
}
