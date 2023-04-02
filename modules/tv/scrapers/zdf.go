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

type Zdf struct {
	ScraperBase
	Timeline cascadia.Matcher
}

func NewZdf(base ScraperBase, timeline string) *Zdf {
	return &Zdf{
		ScraperBase: base,
		Timeline:    cascadia.MustCompile(timeline),
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
	items := cascadia.QueryAll(item, cascadia.MustCompile("li"))
	s.Log.Debug(fmt.Sprintf("#Items: %v", len(items)))
	lastDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.Location)
	for _, i := range items {
		t := cascadia.Query(i, cascadia.MustCompile(".time"))
		n := cascadia.Query(i, cascadia.MustCompile(".overlay-link"))
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
			Time: date,
			Name: name,
		}
	}
}
