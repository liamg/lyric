package genius

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	retry "github.com/avast/retry-go"
)

type songData struct {
	Response songResponse `json:"response"`
}

type songResponse struct {
	Song Song `json:"song"`
}

type Song struct {
	Title  string `json:"title"`
	Artist Artist `json:"primary_artist"`
	URL    string `json:"url"`
	Lyrics string `json:"-"`
}

type Artist struct {
	Title string `json:"name"`
}

func (c *client) GetSong(id int) (*Song, error) {

	songURL := url.URL{
		Scheme: "https",
		Host:   apiHost,
		Path:   fmt.Sprintf("/songs/%d", id),
	}

	data, err := c.get(songURL.String(), true)
	if err != nil {
		return nil, err
	}

	var songData songData
	if err := json.Unmarshal(data, &songData); err != nil {
		return nil, err
	}

	song := songData.Response.Song

	if err := retry.Do(func() error {
		var err error
		song.Lyrics, err = c.scrapeLyrics(song.URL)
		return err
	}, retry.Attempts(3), retry.Delay(time.Second)); err != nil {
		return nil, err
	}

	return &song, nil
}

func (c *client) scrapeLyrics(uri string) (string, error) {
	data, err := c.get(uri, true)
	if err != nil {
		return "", err
	}

	start := `<div class="lyrics">`
	index := strings.Index(string(data), start)
	if index == -1 {
		return "", fmt.Errorf("lyrics unavailable (1)")
	}
	lyrics := string(data)[index+len(start):]

	end := "</div>"
	index = strings.Index(lyrics, end)
	if index == -1 {
		return "", fmt.Errorf("lyrics unavailable (2)")
	}
	lyrics = lyrics[:index]
	return strings.TrimSpace(stripTags(lyrics)), nil
}

func stripTags(s string) string {
	var output string
	var inside bool
	for _, r := range s {
		if inside {
			if r == '>' {
				inside = false
			}
			continue
		}
		if r == '<' {
			inside = true
			continue
		}
		output = output + string(r)
	}

	return output
}
