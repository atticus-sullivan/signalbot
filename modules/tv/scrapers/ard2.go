package scrapers

import (
	"fmt"
	"io"
	"signalbot_go/modules/tv/internal/show"
	"strings"
	"time"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

type Ard2 struct {
	ScraperBase
	Url string
}

func (s *Ard2) Get(now time.Time) (io.ReadCloser, error) {
	url := fmt.Sprintf(s.Url, now.Format("02.01.2006"))
	return s.ScraperBase.Get(url)
}

func (s *Ard2) Parse(r io.ReadCloser, ret chan<- show.Show, now time.Time) {
	defer close(ret)
	root, err := html.Parse(r)
	if err != nil {
		s.Log.Warn(fmt.Sprintf("Error: %v", err))
		return
	}

	items := cascadia.QueryAll(root, cascadia.MustCompile(".accordion-item"))
	s.Log.Debug(fmt.Sprintf("#Items: %v", len(items)))
	lastDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.Location)
	for _, i := range items {
		t := cascadia.Query(i, cascadia.MustCompile(".date"))
		name := cascadia.Query(i, cascadia.MustCompile(".title"))
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
		if date.Before(lastDate) {
			break
		}
		lastDate = date
		ret <- show.Show{
			Time: date,
			Name: strings.Trim(s.node2text(name), " ·"),
		}
	}
}
