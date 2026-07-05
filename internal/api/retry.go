package api

import (
	"bytes"
	"errors"
	"io"
	"math"
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

// isIdempotent returns true for HTTP methods that are safe to auto-retry on a
// 5xx/transport error. PUT is EXCLUDED: Zuora exposes non-idempotent action
// endpoints as PUT (invoice post/reverse/writeoff, order activate/cancel,
// subscription suspend/resume, payment apply) and rejects an Idempotency-Key on
// PUT (HTTP 400), so a retried PUT that already succeeded server-side could
// double-apply. GET/HEAD/OPTIONS/DELETE are naturally idempotent.
func isIdempotent(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions || method == http.MethodDelete
}

// carriesIdempotencyKey reports whether Do attaches an Idempotency-Key to the
// method (POST/PATCH). Single source of truth: Do uses it to decide WHEN to
// attach the key, and the retry layer uses it for the SafeToRetry promise —
// only keyed methods are safe to re-run after a non-retried failure, because
// the key deduplicates a server-side success. PUT is deliberately EXCLUDED:
// Zuora rejects PUT requests carrying an Idempotency-Key with "HTTP 400:
// Request method 'PUT' not supported with Idempotency-Key header" (verified
// against a live tenant), so a failed PUT must NOT be advertised as
// safe-to-retry.
func carriesIdempotencyKey(method string) bool {
	return method == http.MethodPost || method == http.MethodPatch
}

// backoffDuration returns the exponential backoff with equal jitter for a
// 1-based retry attempt: attempt 1 → [1s,1.5s), 2 → [2s,3s), 3 → [4s,6s).
func backoffDuration(attempt int) time.Duration {
	base := time.Duration(1<<(attempt-1)) * time.Second
	return base + time.Duration(rand.Int64N(int64(base/2)))
}

// isTransientBodyCode reports whether a Zuora application error code (the
// numeric "code" of an HTTP-200 success=false reason) is a TRANSIENT condition
// that is safe to retry. Per docs/zuora-api-reference.md the retryable
// categories are codes ending in 50 (locking), 61 (temporary), 70 (limit
// exceeded), and 99 (integration error). Non-numeric codes (e.g. "INVALID")
// are never retried.
func isTransientBodyCode(code string) bool {
	n, err := strconv.Atoi(code)
	if err != nil {
		return false
	}
	switch n % 100 {
	case 50, 61, 70, 99:
		return true
	}
	return false
}

// isRetriableSuccessEnvelope reports whether an HTTP-200 success=false error
// should be retried for the given method. It must be a single-reason transient
// code: parseAPIError leaves Code empty for multi-reason responses, so a batch
// that includes any non-transient reason is conservatively NOT resent. PUT is
// excluded — it carries no Idempotency-Key (Zuora rejects PUT+key), so a resend
// could double-apply; GET/HEAD/OPTIONS/DELETE are idempotent and POST/PATCH
// carry an Idempotency-Key, so their resends deduplicate server-side.
func isRetriableSuccessEnvelope(err error, method string) bool {
	if !isIdempotent(method) && !carriesIdempotencyKey(method) {
		return false
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		return false
	}
	return isTransientBodyCode(apiErr.Code)
}

func (c *Client) doWithRetry(req *http.Request, budget *int) (*http.Response, error) {
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
		// Level-2 verbose: the body is already buffered for retry replay, so
		// logging it here costs no extra read.
		c.vlogBody(">", req.Header.Get("Content-Type"), bodyBytes)
	}

	var lastErr error
	tokenRefreshed := false
	skipBackoff := false

	for attempt := 0; ; attempt++ {
		// Honor cancellation before each attempt. Return the cancellation error
		// (not a stale prior API error) so Ctrl-C is classified as cancellation.
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// Shared attempt budget across BOTH this transport/429/401/5xx loop and
		// Do()'s success-envelope retry loop, so the two cannot multiply: a mixed
		// transient-HTTP + success=false sequence makes at most maxRetries+1 total
		// requests, not (maxRetries+1) squared. Checked before any backoff so an
		// exhausted budget never sleeps first.
		if *budget <= 0 {
			return nil, lastErr
		}

		if attempt > 0 {
			if !skipBackoff {
				total := backoffDuration(attempt)
				c.vlogf("retrying after %.1fs backoff (attempt %d/%d)", total.Seconds(), attempt+1, maxRetries+1)
				if err := c.sleep(ctx, total); err != nil {
					return nil, err
				}
			}
			skipBackoff = false

			if bodyBytes != nil {
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		*budget--
		resp, err := c.httpClient.Do(req)
		if err != nil {
			// A blocked redirect (off-host / cleartext downgrade) is a
			// deterministic policy rejection, not a transient transport error:
			// surface it immediately rather than retrying it (and without the
			// misleading SafeToRetry hint the POST/PATCH branch would add).
			if errors.Is(err, errRedirectRefused) {
				return nil, err
			}
			// Only retry transport errors for idempotent methods. A non-idempotent
			// method may have reached the server, so we do not auto-retry. Mark it
			// safe to re-run only when it carries an Idempotency-Key (POST/PATCH);
			// a keyless PUT could double-apply, so it is NOT advertised as safe.
			if !isIdempotent(req.Method) {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return nil, ctxErr
				}
				c.vlogf("transport error on non-idempotent %s, not retrying: %v", req.Method, err)
				return nil, &APIError{Message: err.Error(), Err: err, SafeToRetry: carriesIdempotencyKey(req.Method), IdemKey: req.Header.Get("Idempotency-Key")}
			}
			c.vlogf("transport error, will retry: %v", err)
			lastErr = err
			continue
		}

		switch {
		case resp.StatusCode == http.StatusTooManyRequests:
			// 429 means the request was rate-limited, not processed. Mutations
			// carry an Idempotency-Key (set in Do) so a replay is safe.
			lastErr = c.readAPIError(resp)
			if d, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
				if d > maxRetryAfter {
					d = maxRetryAfter
				}
				c.vlogf("HTTP 429, honoring Retry-After: %.0fs (cap %.0fs)", d.Seconds(), maxRetryAfter.Seconds())
				if err := c.sleep(ctx, d); err != nil {
					// Cancelled (Ctrl-C) while honoring Retry-After: surface the
					// cancellation, not a stale "HTTP 429" — matching the backoff
					// path above and the loop-top context check.
					return nil, err
				}
				skipBackoff = true
			} else {
				c.vlogf("HTTP 429 rate-limited, retrying with backoff")
			}
			continue

		case resp.StatusCode == http.StatusUnauthorized && !tokenRefreshed:
			// 401 happens before processing (auth is checked first), so a
			// resend after refresh is safe for all methods.
			lastErr = c.readAPIError(resp)
			refreshFn := c.refreshToken
			if refreshFn == nil {
				refreshFn = c.tokenSource
			}
			if refreshFn != nil {
				c.vlogf("HTTP 401, refreshing token and resending")
				token, err := refreshFn(ctx)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Authorization", "Bearer "+token)
				// The literal "Bearer ***" is hardcoded: the refreshed token
				// value must never reach the verbose stream.
				c.vlogf("token refreshed, resending with new credentials (Bearer ***)")
				tokenRefreshed = true
				skipBackoff = true
				continue
			}
			return nil, lastErr

		case resp.StatusCode >= 500 && isIdempotent(req.Method):
			// Only retry 5xx for idempotent methods to avoid duplicate mutations.
			c.vlogf("HTTP %d on idempotent %s, will retry", resp.StatusCode, req.Method)
			lastErr = c.readAPIError(resp)
			continue

		case resp.StatusCode >= 500:
			// Non-idempotent 5xx: do not auto-retry (the mutation may have been
			// applied). Mark safe-to-re-run only for POST/PATCH, which carry an
			// Idempotency-Key; a keyless PUT could double-apply, so leave it unset.
			apiErr := c.readAPIError(resp)
			if ae, ok := apiErr.(*APIError); ok {
				ae.SafeToRetry = carriesIdempotencyKey(req.Method)
				ae.IdemKey = req.Header.Get("Idempotency-Key")
			}
			return nil, apiErr

		default:
			return resp, nil
		}
	}
}

// readAPIError reads and closes the response body, returning a parsed APIError
// that preserves Zuora's real error code/message instead of a generic string.
// The body is also surfaced under level-2 verbose: these retry-layer
// responses (429/401/5xx) never reach Do()'s normal verbose path, yet their
// bodies are the main diagnostic payload for rate-limit and server errors.
func (c *Client) readAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	c.vlogBody("<", resp.Header.Get("Content-Type"), body)
	return parseAPIError(resp.StatusCode, body)
}

// parseRetryAfter parses a Retry-After header, which may be either a number of
// seconds ("120") or an HTTP date ("Wed, 21 Oct 2026 07:28:00 GMT"). It returns
// the wait duration and whether a usable value was found. Negative/past values
// clamp to zero. Seconds large enough to overflow time.Duration (e.g.
// "9223372037") clamp to maxRetryAfter: the multiplication would otherwise
// wrap NEGATIVE, slip past the caller's `d > maxRetryAfter` cap, and turn the
// rate-limit wait into zero-delay hot retries — defeating the very
// hostile-header protection the cap exists for. (The date form is safe:
// time.Until saturates at the Duration limits instead of wrapping.)
func parseRetryAfter(v string) (time.Duration, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(v); err == nil {
		if secs < 0 {
			secs = 0
		}
		if int64(secs) > int64(math.MaxInt64)/int64(time.Second) {
			return maxRetryAfter, true
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
