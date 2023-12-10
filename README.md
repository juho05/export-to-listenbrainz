# export-to-listenbrainz

Upload listen data from various sources to https://listenbrainz.org.

## Sources

- [x] Spotify
- [x] Navidrome

## Installation

First install the [latest release of Go](https://go.dev/dl/) and make sure *GOBIN* is in your *PATH*, then execute the following command in a terminal:
```sh
go install github.com/juho05/export-to-listenbrainz@latest
```

## Usage

The `--token <listenbrainz-token>` flag is necessary to authenticate your requests to [listenbrainz.org](https://listenbrainz.org).

**Example:**
```sh
export-to-listenbrainz --token <listenbrainz-token>
```
where `<listenbrainz-token>` is your *User Token* from https://listenbrainz.org/profile.

If you only want to upload listens in a specific timeframe you can append `--from <YYYY-MM-DD>` and `--until <YYYY-MM-DD` (both inclusive) to any command:

**Examples:**
```sh
# Upload all listens between 15 February 2023 and 17 May 2023 from history1.json and history2.json:
export-to-listenbrainz --token abcd spotify history1.json history2.json --from 2023-02-15 --until 2023-05-17
# Upload all listens between 15 February 2023 and today from https://navidrome.example.com:
export-to-listenbrainz --token abcd navidrome https://navidrome.example.com autht0k€n --from 2023-02-15
# Upload all listens between 01 January 1970 and 15 February 2023 from history.json:
export-to-listenbrainz --token abcd spotify history.json --until 2023-05-17
```

### Spotify

Head over to https://www.spotify.com/account/privacy and request your *extended streaming history*.

After 3-4 weeks you will receive a *ZIP* file with multiple *JSON* files in it.

Now call the `spotify` command of `export-to-listenbrainz` with all of the history files you received.

**Example:** Upload your listening history in `file1.json`, `file2.json` and `file3.json` in the current directory:

```sh
export-to-listenbrainz --token abcd spotify file1.json file2.json file3.json
```

### Navidrome

Sign in to your [Navidrome](https://www.navidrome.org/) instance and open the *Network* tab in the browser dev tools.

Now open the *Songs* tab in Navidrome and select the request to `/api/song` in the dev tools.
Next find the value of the `x-nd-authorization` header and copy it to your clipboard.

With the value of `x-nd-authorization` at hand you can now execute `export-to-listenbrainz`:

**Example:**
```sh
export-to-listenbrainz --token abc navidrome https://navidrome.example.com "Bearer blablablabla"
```
where `Bearer blablablabla` is the value you obtained in the previous step.

## License

Copyright © 2023 Julian Hofmann

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
