package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

func init() {
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, `USAGE: %s [FLAGS] <command> [ARGS]

FLAGS:
  --token <token> (ListenBrainz user token (https://listenbrainz.org/profile/)) REQUIRED
  --cookie <cookiename=cookievalue> (set the cookie for all web requests (except to ListenBrainz))

COMMANDS:
  navidrome <url> <spread-from (YYYY-MM-DD)> <x-nd-authorization value>
  spotify <history-file>

	`, os.Args[0])
	}
}

var cookies []string
var listenBrainzToken string
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
	pflag.Parse()
	if pflag.NArg() == 0 || listenBrainzToken == "" {
		fmt.Println()
		pflag.Usage()
		os.Exit(1)
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
	var err error
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
