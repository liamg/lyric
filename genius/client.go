package genius

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type client struct {
	inner *http.Client
	token AccessToken
}

type metadata struct {
	Status int `json:"status"`
}

func NewClient(accessToken AccessToken) *client {
	return &client{
		token: accessToken,
		inner: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (c *client) get(url string) ([]byte, error) {

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.inner.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with code %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}
