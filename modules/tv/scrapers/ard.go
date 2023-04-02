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

type Ard struct {
	ScraperBase
	Url string
}

func (s *Ard) Get(time.Time) (io.ReadCloser, error) {
	return s.ScraperBase.Get(s.Url)
}

func (s *Ard) Parse(r io.ReadCloser, ret chan<- show.Show, now time.Time) {
	defer close(ret)
	root, err := html.Parse(r)
	if err != nil {
		s.Log.Warn(fmt.Sprintf("Error: %v", err))
		return
	}

	items := cascadia.QueryAll(root, cascadia.MustCompile(".Info-sc-1644nzb-13"))
	s.Log.Debug(fmt.Sprintf("#Items: %v", len(items)))
	lastDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.Location)
	for _, i := range items {
		if i.FirstChild == nil || i.FirstChild.FirstChild == nil || i.FirstChild.NextSibling == nil || i.FirstChild.NextSibling.FirstChild == nil {
			s.Log.Warn(fmt.Sprintf("Error: %v", "failed to parse item, unexpected structure"))
			continue
		}
		tmp := i.FirstChild.FirstChild.Data
		date, err := time.Parse("15:04", strings.TrimSpace(tmp[:len(tmp)-3])) // cut off the suffix " Uhr"
		if err != nil {
			s.Log.Warn(fmt.Sprintf("Error: failed to parse time %v", strings.TrimSpace(tmp[:len(tmp)-3])))
			continue
		}
		date = time.Date(now.Year(), now.Month(), now.Day(), date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), s.Location)
		if date.Before(lastDate) {
			break
		}
		lastDate = date
		name := i.FirstChild.NextSibling.FirstChild.Data

		ret <- show.Show{
			Time: date,
			Name: strings.TrimSpace(name),
		}
	}
}
