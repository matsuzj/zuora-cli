package api

import (
	"bytes"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

const maxRetries = 3

// isIdempotent returns true for HTTP methods that are safe to retry on 5xx.
func isIdempotent(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions || method == http.MethodDelete
}

func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
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
		if attempt > 0 {
			if !skipBackoff {
				backoff := time.Duration(1<<(attempt-1)) * time.Second
				jitter := time.Duration(rand.Int64N(int64(backoff / 2)))
				time.Sleep(backoff + jitter)
			}
			skipBackoff = false

			if bodyBytes != nil {
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Only retry transport errors for idempotent methods
			if !isIdempotent(req.Method) {
				return nil, err
			}
			lastErr = err
			continue
		}

		switch {
		case resp.StatusCode == http.StatusTooManyRequests:
			// 429 is safe to retry for all methods (rate limit, not processed)
			resp.Body.Close()
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if secs, err := strconv.Atoi(retryAfter); err == nil {
					time.Sleep(time.Duration(secs) * time.Second)
					skipBackoff = true
				}
			}
			lastErr = &APIError{StatusCode: resp.StatusCode, Message: "rate limited"}
			continue

		case resp.StatusCode == http.StatusUnauthorized && !tokenRefreshed:
			// 401 token refresh is safe for all methods (not processed yet)
			resp.Body.Close()
			refreshFn := c.refreshToken
			if refreshFn == nil {
				refreshFn = c.tokenSource
			}
			if refreshFn != nil {
				token, err := refreshFn()
				if err != nil {
					return nil, err
				}
				req.Header.Set("Authorization", "Bearer "+token)
				tokenRefreshed = true
				skipBackoff = true
				continue
			}
			return nil, &APIError{StatusCode: resp.StatusCode, Message: "unauthorized"}

		case resp.StatusCode >= 500 && isIdempotent(req.Method):
			// Only retry 5xx for idempotent methods to avoid duplicate mutations
			resp.Body.Close()
			lastErr = &APIError{StatusCode: resp.StatusCode, Message: "server error"}
			continue

		case resp.StatusCode >= 500:
			// Non-idempotent 5xx: return error immediately, don't retry
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			return nil, parseAPIError(resp.StatusCode, body)

		default:
			return resp, nil
		}
	}

	return nil, lastErr
}
