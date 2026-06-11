package api

import (
	"net/http"
)

// Response wraps an HTTP response.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}
