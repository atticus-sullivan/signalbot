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

type SatEins struct {
	ScraperBase
	Url string
}

func (s *SatEins) Get(time.Time) (io.ReadCloser, error) {
	return s.ScraperBase.Get(s.Url)
}

func (s *SatEins) Parse(r io.ReadCloser, ret chan<- show.Show, now time.Time) {
	defer close(ret)
	root, err := html.Parse(r)
	if err != nil {
		s.Log.Warn(fmt.Sprintf("Error: %v", err))
		return
	}

	items := cascadia.QueryAll(root, cascadia.MustCompile("[data-testid=\"epg-teaser-card\"]"))
	s.Log.Debug(fmt.Sprintf("#Items: %v", len(items)))
	lastDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.Location)
	for _, i := range items {
		eles := cascadia.QueryAll(i, cascadia.MustCompile("span"))
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
		if !date.After(lastDate) {
			break
		}
		lastDate = date
		name := strings.TrimSpace(eles[1].FirstChild.Data)

		// parse potential subtitle
		sub := cascadia.Query(i, cascadia.MustCompile("[data-testid=\"teaser-card-meta-info\"]"))
		if sub == nil && len(eles) == 3 {
			s.Log.Warn(fmt.Sprintf("Error: %v", "failed to parse subtitle, unexpected structure"))
			continue
		}
		if sub != nil {
			name += " " + s.node2text(sub)
		}

		retS := show.Show{
			Time: date,
			Name: strings.TrimSpace(name),
		}
		ret <- retS
	}
}
