# lyricli

Song lyrics in your terminal via the [Genius API](https://docs.genius.com).

![screenshot](screenshot.png)

## Usage

Search by song name, artist, or both.

```bash
lyricli "song/artist here"
```

## Install

Install with go:

```bash
go install github.com/liamg/lyricli/cmd/lyricli
```

...or [download the latest binary](https://github.com/liamg/lyricli/releases/latest).

## Configuration

You don't need to do any real configuration, but you'll need a Genius account. Lyricli will pop your browser open to authenticate with Genius on first use, then you're all set. You'll be prompted again if reauthentication is ever needed.
