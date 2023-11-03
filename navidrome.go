package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	req.Header.Set("X-ND-Authorization", pflag.Arg(2))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request song list from Navidrome server: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("request song list from Navidrome server: status: %s", res.Status)
	}

	arrScanner, err := newJsonObjectArrayScanner(res.Body)
	if err != nil {
		return fmt.Errorf("read song list from Navidrome server: %w", err)
	}

	for {
		data, err := arrScanner.nextObject()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read song list from Navidrome server: %w", err)
		}
		type navidromeSong struct {
			PlayCount   int     `json:"playCount"`
			PlayDate    string  `json:"playDate"`
			Title       string  `json:"title"`
			Album       string  `json:"album"`
			Artist      string  `json:"artist"`
			Duration    float32 `json:"duration"`
			DiscNumber  int     `json:"discNumber"`
			TrackNumber int     `json:"trackNumber"`
		}
		var song navidromeSong
		err = json.Unmarshal(data, &song)
		if err != nil {
			return fmt.Errorf("decode song list from Navidrome server: %w", err)
		}
		if song.PlayCount == 0 {
			continue
		}

		playDate, err := time.Parse(time.RFC3339, song.PlayDate)
		if err != nil {
			return fmt.Errorf("decode song list from Navidrome server: parse play date: %w", err)
		}

		for i := 0; i < song.PlayCount; i++ {
			sendListen <- Listen{
				TrackMetadata: TrackMetadata{
					Title:  song.Title,
					Album:  song.Album,
					Artist: song.Artist,
					AdditionalInfo: AdditionalInfo{
						DurationMS:       int(song.Duration * 1000),
						DiscNumber:       song.DiscNumber,
						TrackNumber:      song.TrackNumber,
						SubmissionClient: "navidrome",
					},
				},
				ListenedAt: playDate.Unix(),
			}
		}
	}
	return nil
}
