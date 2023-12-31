package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type AdditionalInfo struct {
	SubmissionClient string `json:"submission_client,omitempty"`
	MusicService     string `json:"music_service,omitempty"`
	OriginURL        string `json:"origin_url,omitempty"`
	SpotifyID        string `json:"spotify_id,omitempty"`
	DiscNumber       int    `json:"discnumber,omitempty"`
	TrackNumber      int    `json:"tracknumber,omitempty"`
	DurationMS       int64  `json:"duration_ms,omitempty"`
}

type TrackMetadata struct {
	Title          string         `json:"track_name"`
	Album          string         `json:"release_name"`
	Artist         string         `json:"artist_name"`
	AdditionalInfo AdditionalInfo `json:"additional_info"`
}

type Listen struct {
	ListenedAt    int64         `json:"listened_at"`
	TrackMetadata TrackMetadata `json:"track_metadata"`
}

func sendListensBatch(listens []Listen) error {
	fmt.Printf("Uploading batch (%d)... ", len(listens))
	type request struct {
		ListenType string   `json:"listen_type"`
		Payload    []Listen `json:"payload"`
	}

	body, err := json.Marshal(request{
		ListenType: "import",
		Payload:    listens,
	})
	if err != nil {
		return fmt.Errorf("encode request body for listenbrainz.org: %w", err)
	}

	for {
		req, err := http.NewRequest(http.MethodPost, "https://api.listenbrainz.org/1/submit-listens", bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("send listens to listenbrainz.org: %w", err)
		}
		req.Header.Add("Authorization", "Token "+listenBrainzToken)
		req.Header.Add("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("send listens to listenbrainz.org: %w", err)
		}
		if res.StatusCode != http.StatusOK {
			body, err := io.ReadAll(res.Body)
			var errorMsg string
			if err == nil {
				errorMsg = string(body)
			}
			res.Body.Close()
			if res.StatusCode == http.StatusTooManyRequests {
				waitSeconds, err := strconv.Atoi(res.Header.Get("X-RateLimit-Reset-In"))
				if err == nil {
					fmt.Printf("Waiting %ds due to rate limit... ", waitSeconds+1)
					time.Sleep(time.Duration(waitSeconds+1) * time.Second)
					continue
				}
			}
			return fmt.Errorf("send listens to listenbrainz.org: status: %s: %s", res.Status, errorMsg)
		}
		res.Body.Close()
		break
	}
	fmt.Println("done")
	return nil
}

func sendListens(listens <-chan Listen) error {
	listensBatch := make([]Listen, 0, 1000)
	for {
		l, ok := <-listens
		if !ok {
			err := sendListensBatch(listensBatch)
			if err != nil {
				return err
			}
			break
		}
		listensBatch = append(listensBatch, l)
		if len(listensBatch) == 1000 {
			err := sendListensBatch(listensBatch)
			if err != nil {
				return err
			}
			listensBatch = listensBatch[:0]
		} else if len(listensBatch) > 1000 {
			panic("len(listensBatch) should never exceed 1000")
		}
	}
	return nil
}
