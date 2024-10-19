package scrapers

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
	"fmt"
	"io"
	"signalbot_go/modules/tv/internal/show"
	"strings"
	"time"
)

type Ard2 struct {
	ScraperBase
	Url string
}

func NewArd2(base ScraperBase, url string) *Ard2 {
	return &Ard2{
		ScraperBase: base,
		Url: url,
	}
}

func (s *Ard2) Get(now time.Time) (io.ReadCloser, error) {
	// url := fmt.Sprintf(s.Url, now.Format("02.01.2006"))
	url := s.Url
	return s.ScraperBase.Get(url)
}

type resp struct {
	Channels []struct {
		Id        string
		TimeSlots [][]struct {
			Id            string
			Title         string
			Subline       string
			BroadcastedOn time.Time
		}
	}
}

func (s *Ard2) Parse(r io.ReadCloser, ret chan<- show.Show, now time.Time) {
	defer close(ret)

	dec := json.NewDecoder(r)
	var resp resp
	if err := dec.Decode(&resp); err != nil {
		s.Log.Warn(fmt.Sprintf("Error: %v", err))
		return
	}

	if len(resp.Channels) != 1 {
		err := errors.New("response should only contain one channel")
		s.Log.Warn(fmt.Sprintf("Error: %v", err))
		return
	}

	seen := make(map[string]struct{})
	for _, i := range resp.Channels[0].TimeSlots {
		for _, j := range i {
			if _, ok := seen[j.Id]; ok {
				continue
			}
			seen[j.Id] = struct{}{}

			var name string
			j.Subline = strings.TrimSpace(j.Subline)
			j.Title = strings.TrimSpace(j.Title)
			if j.Subline != "" {
				name = fmt.Sprintf("%s -- %s", j.Title, j.Subline)
			} else {
				name = j.Title
			}

			ret <- show.Show{
				Date: j.BroadcastedOn,
				Name: name,
			}
		}
	}
}
