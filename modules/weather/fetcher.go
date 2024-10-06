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
}

// validate the fetcher
func (f *Fetcher) Validate() error {
	return nil
}

// get the content from the internet
const baseUrl string = "https://api.open-meteo.com/v1/forecast?"

func (f *Fetcher) getReader(loc Position) (io.ReadCloser, error) {
	params := url.Values{
		"latitude":     {strconv.FormatFloat(float64(loc.Lat), 'f', 6, 32)},
		"longitude":     {strconv.FormatFloat(float64(loc.Lon), 'f', 6, 32)},
		"daily": {"temperature_2m_max", "temperature_2m_min", "precipitation_sum", "weather_code", "sunshine_duration", "wind_speed_10m_max", "wind_direction_10m_dominant", "uv_index_max"},
		"hourly": {"temperature_2m", "precipitation", "snowfall", "weather_code"},
		"current": {"temperature_2m", "relative_humidity_2m", "dew_point_2m", "cloud_cover", "wind_speed_10m", "wind_direction_10m", "wind_gusts_10m", "precipitation", "precipitation_probability", "snowfall", "weather_code"},
		"forecast_days": {"7"},
		"forecast_hours": {"5"},
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
func (f *Fetcher) getFromReader(r io.ReadCloser) (*weatherResp, error) {
	d := json.NewDecoder(r)
	resp := &weatherResp{}
	err := d.Decode(resp)
	if err != nil {
		panic(err)
	}
	return resp, nil
}
