package nowplaying

import (
	"fmt"

	"github.com/godbus/dbus"
)

var players = []string{
	"spotify",
	"ncspot",
}

type Song struct {
	Title  string
	Artist string
}

func GetCurrent() (*Song, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	for _, player := range players {
		object := conn.Object(
			fmt.Sprintf("org.mpris.MediaPlayer2.%s", player),
			"/org/mpris/MediaPlayer2",
		)
		property, err := object.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
		if err != nil {
			continue
		}
		metadata, ok := property.Value().(map[string]dbus.Variant)
		if !ok {
			continue
		}
		var song Song
		if artist, ok := metadata["xesam:artist"]; ok {
			song.Artist = stringFromVariant(artist)
		}
		if title, ok := metadata["xesam:title"]; ok {
			song.Title = stringFromVariant(title)
		}
		if song.Artist == "" && song.Title == "" {
			continue
		}
		return &song, nil
	}
	return nil, fmt.Errorf("cannot communicate with any players")
}

func stringFromVariant(variant dbus.Variant) string {
	if slice, ok := variant.Value().([]string); ok {
		return slice[0]
	}
	if str, ok := variant.Value().(string); ok {
		return str
	}
	return variant.String()
}
