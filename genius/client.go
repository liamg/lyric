package genius

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	retry "github.com/avast/retry-go"
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

func (c *client) get(url string, withRetries bool) ([]byte, error) {

	var data []byte

	attempt := func() error {

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

		resp, err := c.inner.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("request failed with code %d", resp.StatusCode)
		}

		data, err = ioutil.ReadAll(resp.Body)
		return err
	}

	if !withRetries {
		if err := attempt(); err != nil {
			return nil, err
		}
		return data, nil
	}

	if err := retry.Do(attempt, retry.Attempts(3), retry.Delay(time.Second)); err != nil {
		return nil, err
	}

	return data, nil
}
