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
	"fmt"
	"io"
	"signalbot_go/modules/tv/internal/show"
	"strings"
	"time"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

var (
	cascVoxItems cascadia.Matcher = cascadia.MustCompile(".voxde-epg-item")
	cascVoxTime  cascadia.Matcher = cascadia.MustCompile(".time")
	cascVoxTitle cascadia.Matcher = cascadia.MustCompile(".title")
)

type Vox struct {
	ScraperBase
}

func NewVox(base ScraperBase) *Vox {
	return &Vox{
		ScraperBase: base,
	}
}

func (s *Vox) Get(now time.Time) (io.ReadCloser, error) {
	url := "https://www.vox.de/fernsehprogramm/" + now.Format("2006-01-02")
	return s.ScraperBase.Get(url)
}

func (s *Vox) Parse(r io.ReadCloser, ret chan<- show.Show, now time.Time) {
	defer close(ret)
	root, err := html.Parse(r)
	if err != nil {
		s.Log.Warn(fmt.Sprintf("Error: %v", err))
		return
	}

	items := cascadia.QueryAll(root, cascVoxItems)
	s.Log.Debug(fmt.Sprintf("#Items: %v", len(items)))
	for _, i := range items {
		t := cascadia.Query(i, cascVoxTime)
		name := cascadia.Query(i, cascVoxTitle)
		if t == nil || t.FirstChild == nil || name == nil || name.FirstChild == nil {
			s.Log.Warn(fmt.Sprintf("Error: %v", "failed to parse item, unexpected structure"))
			continue
		}
		date, err := time.Parse("15:04", strings.TrimSpace(t.FirstChild.Data))
		if err != nil {
			s.Log.Warn(fmt.Sprintf("Error: failed to parse time %v", strings.TrimSpace(t.FirstChild.Data)))
			continue
		}
		date = time.Date(now.Year(), now.Month(), now.Day(), date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), s.Location)
		ret <- show.Show{
			Date: date,
			Name: strings.TrimSpace(name.FirstChild.Data),
		}
	}
}
