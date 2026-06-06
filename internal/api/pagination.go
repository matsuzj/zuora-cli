package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

const maxPages = 100

// DoPaginated fetches all pages of a paginated API response.
func (c *Client) DoPaginated(method, path string, opts ...RequestOption) ([]json.RawMessage, error) {
	var allData []json.RawMessage

	// A request body (e.g. from `zr api -X POST --paginate --body`) is a
	// single-use io.Reader: the first page drains it, so without buffering,
	// page 2+ would POST an empty body. Materialize it once up front and hand a
	// fresh reader to every page.
	var bodyBytes []byte
	if rc := newRequestConfig(opts); rc.body != nil {
		var err error
		bodyBytes, err = io.ReadAll(rc.body)
		if err != nil {
			return nil, fmt.Errorf("reading request body: %w", err)
		}
	}

	currentPath := path
	for page := 0; page < maxPages; page++ {
		// Pass opts on every page (headers like Zuora-Entity-Ids are needed per-request).
		// buildURL handles absolute nextPage URLs without duplicating query params.
		pageOpts := opts
		if bodyBytes != nil {
			// Copy opts into a fresh slice (so we never alias the caller's
			// backing array) and override the now-drained body with a fresh
			// reader over the buffered bytes. A later WithBody wins because
			// newRequestConfig applies options in order.
			pageOpts = append(append([]RequestOption(nil), opts...), WithBody(bytes.NewReader(bodyBytes)))
		}
		resp, err := c.Do(method, currentPath, pageOpts...)
		if err != nil {
			return nil, err
		}

		var pageResp struct {
			Data     json.RawMessage `json:"data"`
			NextPage string          `json:"nextPage"`
		}
		if err := json.Unmarshal(resp.Body, &pageResp); err != nil {
			// If response is not paginated, return raw body as single element
			allData = append(allData, resp.Body)
			return allData, nil
		}

		if pageResp.Data != nil {
			allData = append(allData, pageResp.Data)
		} else {
			allData = append(allData, resp.Body)
		}

		if pageResp.NextPage == "" {
			break
		}
		currentPath = pageResp.NextPage

		// Safety check: if we've reached the page limit, warn about truncation
		if page == maxPages-1 {
			return allData, fmt.Errorf("pagination limit reached (%d pages); results may be incomplete", maxPages)
		}
	}

	if len(allData) == 0 {
		return nil, fmt.Errorf("no data returned")
	}
	return allData, nil
}
