package main

import (
	"fmt"
	"os"

	"github.com/liamg/lyricli/genius"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lyricli",
	Short: "Display song lyrics via the Genius API",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		token, err := genius.Authenticate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error during authentication: %s\n", err)
			os.Exit(1)
		}

		client := genius.NewClient(token)
		songs, err := client.SearchSongs(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to search: %s\n", err)
			os.Exit(1)

		}

		if len(songs) == 0 {
			fmt.Fprintf(os.Stderr, "Nothing found for '%s'.\n", args[0])
			os.Exit(1)
		}

		song, err := client.GetSong(songs[0].ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to retrieve lyrics: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s (by %s)\n\n", song.Title, song.Artist.Title)
		fmt.Println(song.Lyrics)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	Execute()

}
