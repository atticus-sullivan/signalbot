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
	cascZdfItems cascadia.Matcher = cascadia.MustCompile("li")
	cascZdfTime  cascadia.Matcher = cascadia.MustCompile(".time")
	cascZdfTitle cascadia.Matcher = cascadia.MustCompile(".overlay-link")
)

type Zdf struct {
	ScraperBase
	Timeline cascadia.Matcher
}

func NewZdf(base ScraperBase, timeline cascadia.Matcher) *Zdf {
	return &Zdf{
		ScraperBase: base,
		Timeline:    timeline,
	}
}

func (s *Zdf) Get(time.Time) (io.ReadCloser, error) {
	url := "https://www.zdf.de/live-tv"
	return s.ScraperBase.Get(url)
}

func (s *Zdf) Parse(r io.ReadCloser, ret chan<- show.Show, now time.Time) {
	defer close(ret)
	root, err := html.Parse(r)
	if err != nil {
		s.Log.Warn(fmt.Sprintf("Error: %v", err))
		return
	}

	item := cascadia.Query(root, s.Timeline)
	if item == nil {
		s.Log.Warn(fmt.Sprintf("Error: %v", "failed to parse, timeline not found"))
		return
	}
	items := cascadia.QueryAll(item, cascZdfItems)
	s.Log.Debug(fmt.Sprintf("#Items: %v", len(items)))
	lastDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.Location)
	for _, i := range items {
		t := cascadia.Query(i, cascZdfTime)
		n := cascadia.Query(i, cascZdfTitle)
		if t == nil || t.FirstChild == nil {
			s.Log.Warn(fmt.Sprintf("Error: %v", "failed to parse item, unexpected structure"))
			continue
		}

		date, err := time.Parse("15:04", strings.TrimSpace(strings.Split(t.FirstChild.Data, "-")[0]))
		if err != nil {
			s.Log.Warn(fmt.Sprintf("Error: failed to parse time %v", strings.TrimSpace(strings.Split(t.FirstChild.Data, "-")[0])))
			continue
		}
		date = time.Date(now.Year(), now.Month(), now.Day(), date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), s.Location)

		if date.Before(lastDate) {
			break
		}
		lastDate = date

		name := ""
		for _, attr := range n.Attr {
			if attr.Key == "aria-label" {
				name = attr.Val
			}
		}
		ret <- show.Show{
			Date: date,
			Name: name,
		}
	}
}
