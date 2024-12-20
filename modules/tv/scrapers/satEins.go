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
	"fmt"
	"io"
	"signalbot_go/modules/tv/internal/show"
	"strings"
	"time"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

var (
	cascSatEinsJson cascadia.Matcher = cascadia.MustCompile("#__NEXT_DATA__")

	cascSatEinsItems cascadia.Matcher = cascadia.MustCompile("[data-testid=\"epg-teaser-card\"]")
	cascSatEinsEles  cascadia.Matcher = cascadia.MustCompile("span")
	cascSatEinsSub   cascadia.Matcher = cascadia.MustCompile("[data-testid=\"teaser-card-meta-info\"]")
)

type SatEins struct {
	ScraperBase
	Url string
}

func NewSatEins(base ScraperBase, url string) *SatEins {
	return &SatEins{
		ScraperBase: base,
		Url: url,
	}
}

func (s *SatEins) Get(time.Time) (io.ReadCloser, error) {
	return s.ScraperBase.Get(s.Url)
}

type sat1JsonTeaser struct {
	Title string `json:"title"`
	// Description string `json:"description"`
	StartTime string `json:"startTime"`
	SubTitle  string `json:"subTitle"`
	ShowInfo  string `json:"showInfo"`
}

type sat1JsonBodyEle struct {
	Teasers []sat1JsonTeaser `json:"teasers"`
}

type sat1Json struct {
	Props struct {
		PageProps struct {
			Content struct {
				Body struct {
					Morning   sat1JsonBodyEle `json:"morning"`
					Afternoon sat1JsonBodyEle `json:"afternoon"`
					Evening   sat1JsonBodyEle `json:"evening"`
				} `json:"body"`
			} `json:"content"`
		} `json:"pageProps"`
	} `json:"props"`
}

func (s *SatEins) Parse(r io.ReadCloser, ret chan<- show.Show, now time.Time) {
	defer close(ret)
	root, err := html.Parse(r)
	if err != nil {
		s.Log.Warn(fmt.Sprintf("Error: %v", err))
		return
	}
	if err := s.parseJson(ret, now, root); err != nil {
		s.Log.Info("proceed with fallback (html parsing)")
		_ = s.parseHtml(ret, now, root)
	}
}

func (s *SatEins) parseJsonEle(ele sat1JsonTeaser, now time.Time) (show.Show, error) {
	name := strings.Builder{}
	name.WriteString(ele.Title)
	if ele.SubTitle != "" {
		name.WriteRune(' ')
		name.WriteRune('·')
		name.WriteRune(' ')
		name.WriteString(ele.SubTitle)
	}
	if ele.ShowInfo != "" {
		name.WriteRune(' ')
		name.WriteRune('·')
		name.WriteRune(' ')
		name.WriteString(ele.ShowInfo)
	}

	date, err := time.Parse("15:04", ele.StartTime)
	if err != nil {
		return show.Show{}, fmt.Errorf("failed to parse time %v", ele.StartTime)
	}
	date = time.Date(now.Year(), now.Month(), now.Day(), date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), s.Location)

	retS := show.Show{
		Date: date,
		Name: name.String(),
	}
	return retS, nil
}

func (s *SatEins) parseJson(ret chan<- show.Show, now time.Time, root *html.Node) error {
	jsonNode := cascadia.Query(root, cascSatEinsJson)
	if jsonNode == nil || jsonNode.FirstChild == nil {
		s.Log.Warn(fmt.Sprintf("Error: %v", "failed to get json"))
		return fmt.Errorf("")
	}
	jsonStr := jsonNode.FirstChild.Data
	var jsonStruct sat1Json
	if err := json.Unmarshal([]byte(jsonStr), &jsonStruct); err != nil {
		s.Log.Warn(fmt.Sprintf("Error: %v", "failed to parse json"))
		return fmt.Errorf("")
	}

	for _, ele := range jsonStruct.Props.PageProps.Content.Body.Morning.Teasers {
		retS, err := s.parseJsonEle(ele, now)
		if err != nil {
			s.Log.Warn(fmt.Sprintf("Error: %v", err))
			continue
		}
		ret <- retS
	}
	for _, ele := range jsonStruct.Props.PageProps.Content.Body.Afternoon.Teasers {
		retS, err := s.parseJsonEle(ele, now)
		if err != nil {
			s.Log.Warn(fmt.Sprintf("Error: %v", err))
			continue
		}
		ret <- retS
	}
	for _, ele := range jsonStruct.Props.PageProps.Content.Body.Evening.Teasers {
		retS, err := s.parseJsonEle(ele, now)
		if err != nil {
			s.Log.Warn(fmt.Sprintf("Error: %v", err))
			continue
		}
		ret <- retS
	}
	return nil
}

func (s *SatEins) parseHtml(ret chan<- show.Show, now time.Time, root *html.Node) error {
	items := cascadia.QueryAll(root, cascSatEinsItems) // missing "epg-live-teaser" but this only contains "bis HH:MM", no start date
	s.Log.Debug(fmt.Sprintf("#Items: %v", len(items)))
	// lastDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.Location)
	for _, i := range items {
		eles := cascadia.QueryAll(i, cascSatEinsEles)
		if len(eles) < 2 || len(eles) > 3 || eles[0].FirstChild == nil || eles[1].FirstChild == nil {
			s.Log.Warn(fmt.Sprintf("Error: %v", "failed to parse item, unexpected structure"))
			continue
		}
		date, err := time.Parse("15:04", strings.TrimSpace(eles[0].FirstChild.Data))
		if err != nil {
			s.Log.Warn(fmt.Sprintf("Error: failed to parse time %v", strings.TrimSpace(eles[0].FirstChild.Data)))
			continue
		}
		date = time.Date(now.Year(), now.Month(), now.Day(), date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), s.Location)
		// if !date.After(lastDate) {
		// 	break
		// }
		// lastDate = date
		name := strings.TrimSpace(eles[1].FirstChild.Data)

		// parse potential subtitle
		sub := cascadia.Query(i, cascSatEinsSub)
		if sub == nil && len(eles) == 3 {
			s.Log.Warn(fmt.Sprintf("Error: %v", "failed to parse subtitle, unexpected structure"))
			continue
		}
		if sub != nil {
			name += " " + s.node2text(sub)
		}

		retS := show.Show{
			Date: date,
			Name: strings.TrimSpace(name),
		}
		ret <- retS
	}
	return nil
}
