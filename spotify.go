package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

func spotify(sendListen chan<- Listen) error {
	if pflag.NArg() < 2 {
		pflag.Usage()
		os.Exit(1)
	}
	for i := 1; i < pflag.NArg(); i++ {
		file, err := os.Open(pflag.Arg(i))
		fmt.Printf("Exporting %s...\n", file.Name())
		if err != nil {
			return fmt.Errorf("open spotify history file: %w", err)
		}
		defer file.Close()

		type spotifyListen struct {
			TS              string `json:"ts"`
			MSPlayed        int64  `json:"ms_played"`
			TrackName       string `json:"master_metadata_track_name"`
			AlbumName       string `json:"master_metadata_album_album_name"`
			ArtistName      string `json:"master_metadata_album_artist_name"`
			SpotifyTrackURI string `json:"spotify_track_uri"`
		}
		arrScanner, err := newJsonObjectArrayScanner[spotifyListen](file)
		if err != nil {
			return fmt.Errorf("read song list from history file: %w", err)
		}
		for {
			song, err := arrScanner.nextObject()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("read song list from history file: %w", err)
			}

			listenedAt, err := time.Parse(time.RFC3339, song.TS)
			if err != nil {
				return fmt.Errorf("read song list from history file: invalid timestamp '%s': %w", song.TS, err)
			}

			if song.TrackName == "" || song.ArtistName == "" || song.MSPlayed < 30000 {
				continue
			}
			listenedAtDate := time.Date(listenedAt.Year(), listenedAt.Month(), listenedAt.Day(), 0, 0, 0, 0, listenedAt.Location())
			if listenedAtDate.Compare(from) < 0 || listenedAtDate.Compare(until) > 0 {
				continue
			}

			songURL := "https://open.spotify.com/track/" + strings.Split(song.SpotifyTrackURI, ":")[2]

			sendListen <- Listen{
				TrackMetadata: TrackMetadata{
					Title:  song.TrackName,
					Album:  song.AlbumName,
					Artist: song.ArtistName,
					AdditionalInfo: AdditionalInfo{
						MusicService:     "spotify.com",
						SubmissionClient: "export-to-listenbrainz",
						SpotifyID:        songURL,
						OriginURL:        songURL,
					},
				},
				ListenedAt: listenedAt.Unix(),
			}
		}
		file.Close()
	}
	return nil
}
