package api

import (
	"io"
	"net/url"
)

type requestConfig struct {
	headers      map[string]string
	query        url.Values
	body         io.Reader
	checkSuccess bool
}

// RequestOption configures an API request.
type RequestOption func(*requestConfig)

// WithHeader adds a custom header to the request.
func WithHeader(key, value string) RequestOption {
	return func(rc *requestConfig) {
		rc.headers[key] = value
	}
}

// WithQuery adds a query parameter to the request.
func WithQuery(key, value string) RequestOption {
	return func(rc *requestConfig) {
		rc.query.Set(key, value)
	}
}

// WithQuerySlice adds multiple values for a single query key (e.g. filter[]).
func WithQuerySlice(key string, values []string) RequestOption {
	return func(rc *requestConfig) {
		for _, v := range values {
			rc.query.Add(key, v)
		}
	}
}

// WithBody sets the request body.
func WithBody(r io.Reader) RequestOption {
	return func(rc *requestConfig) {
		rc.body = r
	}
}

// WithCheckSuccess enables Zuora success flag checking on the response.
// When enabled, HTTP 200 responses with {"success": false} are treated as errors.
func WithCheckSuccess() RequestOption {
	return func(rc *requestConfig) {
		rc.checkSuccess = true
	}
}

func newRequestConfig(opts []RequestOption) *requestConfig {
	rc := &requestConfig{
		headers: make(map[string]string),
		query:   make(url.Values),
	}
	for _, opt := range opts {
		opt(rc)
	}
	return rc
}
