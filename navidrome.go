package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

func navidrome(sendListen chan<- Listen) error {
	if pflag.NArg() != 3 {
		pflag.Usage()
		os.Exit(1)
	}

	req, err := newHTTPRequest(http.MethodGet, strings.TrimSuffix(pflag.Arg(1), "/")+"/api/song", nil)
	if err != nil {
		return fmt.Errorf("create song list request: %w", err)
	}
	req.Header.Set("X-ND-Authorization", pflag.Arg(3))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request song list from Navidrome server: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("request song list from Navidrome server: status: %s", res.Status)
	}

	type navidromeSong struct {
		PlayCount   int     `json:"playCount"`
		Title       string  `json:"title"`
		Album       string  `json:"album"`
		Artist      string  `json:"artist"`
		Duration    float32 `json:"duration"`
		DiscNumber  int     `json:"discNumber"`
		TrackNumber int     `json:"trackNumber"`
	}
	arrScanner, err := newJsonObjectArrayScanner[navidromeSong](res.Body)
	if err != nil {
		return fmt.Errorf("read song list from Navidrome server: %w", err)
	}

	for {
		song, err := arrScanner.nextObject()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read song list from Navidrome server: %w", err)
		}
		if song.PlayCount == 0 {
			continue
		}

		for i := 0; i < song.PlayCount; i++ {
			sendListen <- Listen{
				TrackMetadata: TrackMetadata{
					Title:  song.Title,
					Album:  song.Album,
					Artist: song.Artist,
					AdditionalInfo: AdditionalInfo{
						DiscNumber:       song.DiscNumber,
						TrackNumber:      song.TrackNumber,
						MusicService:     "navidrome",
						SubmissionClient: "export-to-listenbrainz",
						DurationMS:       int64(song.Duration),
					},
				},
				ListenedAt: from.Add(time.Duration(rand.Int63n(int64(until.Add(23*time.Hour + 59*time.Minute + 59*time.Second).Sub(from))))).Unix(),
			}
		}
	}
	return nil
}
