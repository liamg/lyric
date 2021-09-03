package genius

import (
	"encoding/json"
	"net/url"
)

type searchResponse struct {
	Metadata   metadata   `json:"meta"`
	SearchData searchData `json:"response"`
}

type searchData struct {
	Hits []hit `json:"hits"`
}

type SearchResultType string

const (
	SearchResultTypeSong SearchResultType = "song"
)

type hit struct {
	Type   SearchResultType `json:"type"`
	Result SearchResult     `json:"result"`
}

type SearchResult struct {
	ID   int    `json:"id"`
	Text string `json:"full_title"`
}

func (c *client) SearchSongs(term string) ([]SearchResult, error) {
	searchURL := url.URL{
		Scheme: "https",
		Host:   apiHost,
		Path:   "/search",
	}
	q := searchURL.Query()
	q.Set("q", term)
	searchURL.RawQuery = q.Encode()

	data, err := c.get(searchURL.String(), true)
	if err != nil {
		return nil, err
	}

	var resp searchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, result := range resp.SearchData.Hits {
		results = append(results, result.Result)
	}
	return results, nil
}
