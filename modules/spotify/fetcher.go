package spotify

// signalbot
// Copyright (C) 2024  Lukas Heindl
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/jellydator/ttlcache/v3"
)

var (
	ErrNetwork           error = errors.New("Error retreiving from network")
	ErrInvalidResultCode error = errors.New("Error invalid result code")
	ErrPostStatusCode    error = errors.New("Error when posting query, statuscode not OK")
)

type album struct {
	Title    string
	CntSongs uint
	Type     string
	Artist   string
	Release  string
}

func (b album) String() string {
	builder := strings.Builder{}

	builder.WriteString(b.Artist)
	builder.WriteString(" -> ")
	builder.WriteString(b.Title)

	builder.WriteString(" [")
	builder.WriteString(strconv.Itoa(int(b.CntSongs)))
	builder.WriteString("] ")

	builder.WriteString(b.Type)
	builder.WriteString(" | ")
	builder.WriteString(b.Release)

	return builder.String()
}

func (b album) AddString() string {
	return "> " + b.String()
}
func (b album) RemString() string {
	return "< " + b.String()
}
func (b album) Equals(o album) bool {
	return b == o
}

type albumItem []album

func (bs albumItem) String() string {
	builder := strings.Builder{}

	first := true
	for _, b := range bs {
		if !first {
			builder.WriteRune('\n')
			builder.WriteRune('\n')
		} else {
			first = false
		}
		builder.WriteString(b.String())
		builder.WriteRune('\n')
	}

	return builder.String()
}

type accessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// fetches stuff. Maybe some day this will have data members (e.g. if caching
// is implemented)
// Has to be instanciated via `NewFetcher`
type Fetcher struct {
	cache        *ttlcache.Cache[string, accessToken]
	log          *slog.Logger
	clientId     string
	clientSecret string
	size         uint
}

func NewFetcher(log *slog.Logger, querySize uint, clientId string, clientSecret string) *Fetcher {
	f := &Fetcher{
		cache:        ttlcache.New(ttlcache.WithTTL[string, accessToken](50*time.Minute), ttlcache.WithDisableTouchOnHit[string, accessToken]()),
		log:          log,
		clientId:     clientId,
		clientSecret: clientSecret,
		size:         querySize,
	}
	return f
}

func (f *Fetcher) auth() (*accessToken, error) {
	if v := f.cache.Get("all"); v != nil && !v.IsExpired() {
		tmp := v.Value()
		return &tmp, nil
	}

	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", bytes.NewBufferString("grant_type=client_credentials"))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(f.clientId+":"+f.clientSecret)))

	respS, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer respS.Body.Close()

	if respS.StatusCode != http.StatusOK {
		return nil, ErrNetwork
	}

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, respS.Body)
	if err != nil {
		return nil, err
	}

	resp := accessToken{}
	err = json.Unmarshal(buf.Bytes(), &resp)
	if err != nil {
		return nil, err
	}

	if resp.TokenType != "Bearer" {
		return nil, ErrInvalidResultCode
	}

	f.cache.Set("all", resp, ttlcache.DefaultTTL)
	return &resp, nil
}

func (f *Fetcher) get(artistIds []string) (albumItem, error) {
	jwt, err := f.auth()
	if err != nil {
		return nil, err
	}

	var ret albumItem

	for _, artistId := range artistIds {
		c := make(chan album, 5)
		go f.getStep(artistId, jwt, int(f.size), c)
		for b := range c {
			ret = append(ret, b)
		}
	}
	f.log.Debug("total amount of items fetched", slog.Int("#", len(ret)))
	return ret, nil
}

type spotifyResponse struct {
	Next  string `json:"next"`
	Total uint   `json:"total"`
	Items []struct {
		AlbumType   string `json:"album_type"`
		TotalTracks uint   `json:"total_tracks"`
		Name        string `json:"name"`
		ReleaseDate string `json:"release_date"`
		AlbumGroup  string `json:"album_group"`
		Artists []struct{
			Name string `json:"name"`
		} `json:"artists"`
	} `json:"items"`
}

func (f *Fetcher) getStep(artistId string, jwt *accessToken, size int, out chan<- album) {
	defer close(out)

	offset := 0
	running := true
	for ; running; offset += size {
		r, err := f.getReader(artistId, jwt, offset, size)
		if err != nil {
			f.log.Warn(err.Error())
			return
		}

		buf := &bytes.Buffer{}
		_, err = io.Copy(buf, r)
		f.log.Debug("Query", "response", buf)
		if err != nil {
			f.log.Warn(err.Error())
		}

		resp := spotifyResponse{}
		err = json.Unmarshal(buf.Bytes(), &resp)
		if err != nil {
			f.log.Warn(err.Error())
		}

		for _, j := range resp.Items {
			b := album{
				Title:    j.Name,
				CntSongs: j.TotalTracks,
				Type:     j.AlbumType,
				Release:  j.ReleaseDate,
			}
			builder := strings.Builder{}
			first := true
			for _, a := range j.Artists {
				if !first {
					builder.WriteString(" + ")
				}
				builder.WriteString(a.Name)
			}
			b.Artist = builder.String()

			out <- b
		}

		if len(resp.Items)+offset >= int(resp.Total) {
			if len(resp.Items)+offset > int(resp.Total) {
				f.log.Warn("unexpected len vs totalResults", slog.Int("len", len(resp.Items)+offset), slog.Int("total", int(resp.Total)))
			}
			running = false
		}
		// f.log.Debug("Read items", slog.Int("count", len(resp.Result.Articles)+offset), slog.Any("query", q))
	}
}

// get the content from the internet
func (f *Fetcher) getReader(artistId string, jwt *accessToken, offset int, size int) (io.ReadCloser, error) {
	var err error
	if jwt == nil {
		jwt, err = f.auth()
		if err != nil {
			return nil, err
		}
	}

	v := url.Values{}
	v.Set("market", "DE")
	v.Set("include_groups", "album,single")
	v.Set("offset", strconv.Itoa(offset))
	v.Set("limit", strconv.Itoa(size))

	url := "https://api.spotify.com/v1/artists/"+artistId+"/albums?"+v.Encode()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	f.log.Debug("query ", "url", url)
	if err != nil {
		return nil, err
	}
	f.log.Debug("query ", "auth", jwt.AccessToken)
	req.Header.Set("Authorization", "Bearer "+jwt.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, ErrPostStatusCode
	}
	return resp.Body, nil
}
