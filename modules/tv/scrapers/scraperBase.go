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
	"net/http"
	"strings"
	"time"

	"log/slog"
	"golang.org/x/net/html"
)

type ScraperBase struct {
	NameS    string
	Location *time.Location
	Log      *slog.Logger
	File     string
}

func NewScraperBase(base *slog.Logger, name string, loc *time.Location) ScraperBase {
	s := ScraperBase{
		Log:      base.With(slog.String("scraper", name)),
		NameS:    name,
		File:     name + ".html",
		Location: loc,
	}
	return s
}

func (s *ScraperBase) Name() string {
	return s.NameS
}

func (s *ScraperBase) Get(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		s.Log.Warn(fmt.Sprintf("Error: %v", err))
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("fetching url: %v failed with status: %s", url, resp.Status)
		s.Log.Warn(fmt.Sprintf("Error: %v", err))
		return nil, err
	}
	return resp.Body, nil
}

// func (s *ScraperBase) GetFromFile() (io.ReadCloser, error){
// 	f,err := os.Open(s.File)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return f, nil
// }

func (s *ScraperBase) TimeToday(timeT time.Time, today time.Time) time.Time {
	return time.Date(today.Year(), today.Month(), today.Day(), timeT.Hour(), timeT.Minute(), timeT.Second(), timeT.Nanosecond(), s.Location)
}

func (s *ScraperBase) node2text(node *html.Node) string {
	dat := strings.TrimSpace(node.Data)
	if node.Data == "style" || dat == "" {
		return ""
	}
	if node.Type == html.TextNode {
		return "· " + dat + " "
	}

	builder := strings.Builder{}
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		builder.WriteString(s.node2text(n))
	}
	return builder.String()
}
