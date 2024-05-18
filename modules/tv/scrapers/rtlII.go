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
	cascRtlIIItems cascadia.Matcher = cascadia.MustCompile("epg-broadcast-row")
	cascRtlIITitle cascadia.Matcher = cascadia.MustCompile(".teaser-title")
)

type Rtl2 struct {
	ScraperBase
}

func (s *Rtl2) Get(time.Time) (io.ReadCloser, error) {
	url := "https://www.rtl2.de/tv-programm/" + time.Now().Format("2006-01-02")
	return s.ScraperBase.Get(url)
}

func (s *Rtl2) Parse(r io.ReadCloser, ret chan<- show.Show, now time.Time) {
	defer close(ret)
	root, err := html.Parse(r)
	if err != nil {
		s.Log.Warn(fmt.Sprintf("Error: %v", err))
		return
	}

	items := cascadia.QueryAll(root, cascRtlIIItems)
	s.Log.Debug(fmt.Sprintf("#Items: %v", len(items)))
	for _, i := range items {
		var date time.Time
		for _, attr := range i.Attr {
			if attr.Key == "start" {
				date, err = time.Parse("2006-01-02T15:04:05Z07:00", attr.Val)
			}
		}
		if err != nil || date.IsZero() {
			s.Log.Warn(fmt.Sprintf("Error: failed to parse time or none provided %v", date))
			continue
		}

		name := strings.Builder{}
		first := true
		for _, t := range cascadia.QueryAll(i, cascRtlIITitle) {
			if t.FirstChild == nil {
				s.Log.Warn(fmt.Sprintf("Error: %v", "failed to parse item, unexpected structure"))
				continue
			}
			if !first {
				name.WriteString(" · ")
			} else {
				first = false
			}
			name.WriteString(strings.TrimSpace(t.FirstChild.Data))
		}

		retS := show.Show{
			Date: date,
			Name: strings.TrimSpace(name.String()),
		}
		ret <- retS
	}
}
