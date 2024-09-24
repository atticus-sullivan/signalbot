package weather

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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

var (
	ErrNetwork error = errors.New("Error retreiving from network")
	ErrLang    error = errors.New("Invalid language selected")
	ErrUnit    error = errors.New("Invalid unitsystem selected")
	ErrKey     error = errors.New("Invalid API-key. Expected to have 32 chars")
)

// fetches stuff. (e.g. if caching might be implemented at this level)
// data members are only public to be able to (un)marshal them
type Fetcher struct {
	ApiKey     string `yaml:"openweatherKey"`
	Lang       string `yaml:"lang"`
	Unitsystem string `yaml:"unitsystem"`
}

// validate the fetcher
func (f *Fetcher) Validate() error {
	if v, ok := langs[f.Lang]; !ok || !v {
		return ErrLang
	}
	if f.Unitsystem != "metric" && f.Unitsystem != "imperial" && f.Unitsystem != "standard" {
		return ErrUnit
	}
	if len(f.ApiKey) != 32 {
		return ErrKey
	}
	return nil
}

// get the content from the internet
const baseUrl string = "https://api.openweathermap.org/data/2.5/onecall?"

func (f *Fetcher) getReader(loc Position) (io.ReadCloser, error) {
	params := url.Values{
		"exclude": {"minutely"},
		"appid":   {f.ApiKey},
		"lat":     {strconv.FormatFloat(float64(loc.Lat), 'f', 4, 32)},
		"lon":     {strconv.FormatFloat(float64(loc.Lon), 'f', 4, 32)},
		"units":   {f.Unitsystem},
		"lang":    {f.Lang},
	}
	resp, err := http.Get(baseUrl + params.Encode())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, ErrNetwork
	}
	return resp.Body, nil
}

// parse the content from an arbitrary reader (can be a file, a network
// response body or something else)
func (f *Fetcher) getFromReader(r io.ReadCloser) (*openweatherResp, error) {
	d := json.NewDecoder(r)
	resp := &openweatherResp{}
	err := d.Decode(resp)
	if err != nil {
		panic(err)
	}
	return resp, nil
}
