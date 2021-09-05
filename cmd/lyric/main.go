package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/liamg/lyric/genius"
	"github.com/liamg/lyric/nowplaying"
	"github.com/liamg/tml"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lyric [song name]",
	Short: "Display song lyrics via the Genius API",
	Run: func(cmd *cobra.Command, args []string) {

		if noColours {
			tml.DisableFormatting()
		}

		var term string

		if len(args) == 0 {
			current, err := nowplaying.GetCurrent()
			if err != nil {
				fail("No search terms provided, and could not determine currently playing song.")
			}
			term = fmt.Sprintf("%s %s", current.Title, current.Artist)
		} else {
			term = strings.Join(args, " ")
		}

		token, err := genius.Authenticate()
		if err != nil {
			fail("Authentication failed: %s", err)
		}

		client := genius.NewClient(token)
		songs, err := client.SearchSongs(term)
		if err != nil {
			fail("Failed to search: %s", err)
		}

		if len(songs) == 0 {
			fail("Nothing found for '%s'.", term)
		}

		song, err := client.GetSong(songs[0].ID)
		if err != nil {
			fail("Failed to retrieve lyrics: %s", err)
		}

		printLyrics(song)
	},
}

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

var colourSequence = []string{
	"blue",
	"yellow",
	"green",
	"white",
}

func printLyrics(song *genius.Song) {

	var output string

	output += tml.Sprintf(
		"\n<white><bold>%s</bold></white>\n<dim>%s\n\n",
		song.Title,
		song.Artist.Title,
	)

	for i, verse := range song.Lyrics.Verses {
		if verse.Label != "" {
			output += tml.Sprintf("<dim>[%s]\n", verse.Label)
		}

		for _, line := range verse.Lines {
			format := fmt.Sprintf("<%s>", colourSequence[i%len(colourSequence)])
			output += tml.Sprintf(format+"%s\n", line)
		}

		output += "\n"
	}

	pageOutput(output)
}

func pageOutput(raw string) {

	pager := os.Getenv("PAGER")
	if pager == "" || noPaging {
		fmt.Println(raw)
		return
	}

	// Could read $PAGER rather than hardcoding the path.
	cmd := exec.Command(pager)

	// Feed it with the string you want to display.
	cmd.Stdin = strings.NewReader(raw)

	// This is crucial - otherwise it will write to a null device.
	cmd.Stdout = os.Stdout

	// Fork off a process and wait for it to terminate.
	err := cmd.Run()
	if err != nil {
		fmt.Println(raw)
	}

}

var noColours bool
var noPaging bool

func main() {
	rootCmd.Flags().BoolVarP(&noColours, "disable-ansi", "n", noColours, "Disable ANSI colours/formatting")
	rootCmd.Flags().BoolVarP(&noPaging, "disable-paging", "d", noPaging, "Disable paging of output")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
