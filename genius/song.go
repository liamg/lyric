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
	Lyrics Lyrics `json:"-"`
}

type Lyrics struct {
	Verses []Verse `json:"-"`
}

type Verse struct {
	Label string
	Lines []string
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

	var rawLyrics string
	if err := retry.Do(func() error {
		var err error
		rawLyrics, err = c.scrapeLyrics(song.URL)
		return err
	}, retry.Attempts(3), retry.Delay(time.Second)); err != nil {
		return nil, err
	}

	song.Lyrics = parseLyrics(rawLyrics)

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

func parseLyrics(raw string) Lyrics {
	var lyrics Lyrics
	var verse Verse

	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if len(verse.Lines) > 0 && line == "" {
			lyrics.Verses = append(lyrics.Verses, verse)
			verse = Verse{}
			continue
		}
		if len(verse.Lines) == 0 && verse.Label == "" && strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			verse.Label = line[1 : len(line)-1]
			continue
		}
		verse.Lines = append(verse.Lines, line)
	}

	if len(verse.Lines) > 0 {
		lyrics.Verses = append(lyrics.Verses, verse)
	}

	return lyrics
}
