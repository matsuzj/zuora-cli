package api

import (
	"encoding/json"
	"net/http"
)

// Response wraps an HTTP response.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// JSON unmarshals the response body into v.
func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// String returns the response body as a string.
func (r *Response) String() string {
	return string(r.Body)
}
