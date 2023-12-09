package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

func init() {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, `USAGE: %s [FLAGS] <command> [ARGS]

FLAGS:
  --token <token> (ListenBrainz user token (https://listenbrainz.org/profile/)) REQUIRED
  --cookie <cookiename=cookievalue> (set the cookie for all web requests (except to ListenBrainz))
  --from <YYYY-MM-DD> (only export listens >= this date)
  --until <YYYY-MM-DD> (only export listens <= this date)

COMMANDS:
  navidrome <url> <x-nd-authorization value>
  spotify <history-files...>

	`, os.Args[0])
	}
}

var cookies []string
var listenBrainzToken string
var from = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
var until = time.Now().UTC()
var uploadDone = make(chan struct{})

func newHTTPRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}
	for _, cookie := range cookies {
		parts := strings.Split(cookie, "=")
		if len(parts) == 1 {
			parts = append(parts, "")
		}
		req.AddCookie(&http.Cookie{
			Name:  parts[0],
			Value: strings.Join(parts[1:], "="),
		})
	}
	return req, nil
}

func main() {
	pflag.StringArrayVar(&cookies, "cookie", nil, "set a cookie for all web requests (except to Listenbrainz)")
	pflag.StringVar(&listenBrainzToken, "token", "", "ListenBrainz user token (https://listenbrainz.org/profile)")
	var fromStr string
	pflag.StringVar(&fromStr, "from", "", "only export listens >= this date")
	var untilStr string
	pflag.StringVar(&untilStr, "until", "", "only export listens <= this date")
	pflag.Parse()
	if pflag.NArg() == 0 || listenBrainzToken == "" {
		pflag.Usage()
		os.Exit(1)
	}

	var err error
	if fromStr != "" {
		from, err = time.Parse(time.DateOnly, fromStr)
		if err != nil {
			pflag.Usage()
			os.Exit(1)
		}
	}
	if untilStr != "" {
		until, err = time.Parse(time.DateOnly, untilStr)
		if err != nil {
			pflag.Usage()
			os.Exit(1)
		}
	}

	listenChannel := make(chan Listen)
	go func() {
		err := sendListens(listenChannel)
		if err != nil {
			close(listenChannel)
			fmt.Fprintln(os.Stderr, "ERROR:", err)
			os.Exit(1)
		}
		close(uploadDone)
	}()
	switch pflag.Arg(0) {
	case "navidrome":
		err = navidrome(listenChannel)
	case "spotify":
		err = spotify(listenChannel)
	default:
		err = fmt.Errorf("unknown source: %s", pflag.Arg(0))
	}
	close(listenChannel)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
	<-uploadDone
}
