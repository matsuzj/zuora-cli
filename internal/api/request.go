package api

import "io"

type requestConfig struct {
	headers map[string]string
	query   map[string]string
	body    io.Reader
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
		rc.query[key] = value
	}
}

// WithBody sets the request body.
func WithBody(r io.Reader) RequestOption {
	return func(rc *requestConfig) {
		rc.body = r
	}
}

func newRequestConfig(opts []RequestOption) *requestConfig {
	rc := &requestConfig{
		headers: make(map[string]string),
		query:   make(map[string]string),
	}
	for _, opt := range opts {
		opt(rc)
	}
	return rc
}
