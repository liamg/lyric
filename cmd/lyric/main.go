package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/liamg/lyric/genius"
	"github.com/liamg/tml"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lyric [song name]",
	Short: "Display song lyrics via the Genius API",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		if noColours {
			tml.DisableFormatting()
		}

		term := strings.Join(args, " ")

		token, err := genius.Authenticate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error during authentication: %s\n", err)
			os.Exit(1)
		}

		client := genius.NewClient(token)
		songs, err := client.SearchSongs(term)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to search: %s\n", err)
			os.Exit(1)

		}

		if len(songs) == 0 {
			fmt.Fprintf(os.Stderr, "Nothing found for '%s'.\n", term)
			os.Exit(1)
		}

		song, err := client.GetSong(songs[0].ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to retrieve lyrics: %s\n", err)
			os.Exit(1)
		}

		printLyrics(song)

	},
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
