package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type AdditionalInfo struct {
	DurationMS       int    `json:"duration_ms,omitempty"`
	SubmissionClient string `json:"submission_client,omitempty"`
	MusicService     string `json:"music_service,omitempty"`
	OriginURL        string `json:"origin_url,omitempty"`
	SpotifyID        string `json:"spotify_id,omitempty"`
	DiscNumber       int    `json:"discnumber,omitempty"`
	TrackNumber      int    `json:"tracknumber,omitempty"`
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
	fmt.Printf("Uploading batch (%d)...\n", len(listens))
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
		res.Body.Close()
		if res.StatusCode != http.StatusOK {
			if res.StatusCode == http.StatusTooManyRequests {
				waitSeconds, err := strconv.Atoi(res.Header.Get("X-RateLimit-Reset-In"))
				if err == nil {
					time.Sleep(time.Duration(waitSeconds+1) * time.Second)
					continue
				}
			}
			return fmt.Errorf("send listens to listenbrainz.org: status: %s", res.Status)
		}
		break
	}
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
		if len(listensBatch) < 1000 {
			listensBatch = append(listensBatch, l)
		}
		if len(listensBatch) == 1000 {
			err := sendListensBatch(listensBatch)
			if err != nil {
				return err
			}
			listensBatch = listensBatch[:0]
		}
	}
	return nil
}
