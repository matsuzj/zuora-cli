package api

import (
	"encoding/json"
	"fmt"
)

const maxPages = 100

// DoPaginated fetches all pages of a paginated API response.
func (c *Client) DoPaginated(method, path string, opts ...RequestOption) ([]json.RawMessage, error) {
	var allData []json.RawMessage

	currentPath := path
	for page := 0; page < maxPages; page++ {
		// Pass opts on every page (headers like Zuora-Entity-Ids are needed per-request).
		// buildURL handles absolute nextPage URLs without duplicating query params.
		resp, err := c.Do(method, currentPath, opts...)
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
