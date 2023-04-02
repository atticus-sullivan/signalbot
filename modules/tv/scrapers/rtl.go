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

type Rtl struct {
	ScraperBase
}

func (s *Rtl) Get(time.Time) (io.ReadCloser, error) {
	url := "https://www.rtl.de/fernsehprogramm/" + time.Now().Format("2006-01-02")
	return s.ScraperBase.Get(url)
}

func (s *Rtl) Parse(r io.ReadCloser, ret chan<- show.Show, now time.Time) {
	defer close(ret)
	root, err := html.Parse(r)
	if err != nil {
		s.Log.Warn(fmt.Sprintf("Error: %v", err))
		return
	}

	items := cascadia.QueryAll(root, cascadia.MustCompile(".rtlde-epg-item"))
	s.Log.Debug(fmt.Sprintf("#Items: %v", len(items)))
	for _, i := range items {
		node := cascadia.Query(i, cascadia.MustCompile(".title"))
		if node == nil || node.FirstChild == nil || node.FirstChild.NextSibling == nil || node.FirstChild.NextSibling.NextSibling == nil {
			s.Log.Warn(fmt.Sprintf("Error: %v", "failed to parse item, unexpected structure"))
			continue
		}
		date, err := time.Parse("15:04", strings.TrimSpace(node.FirstChild.Data))
		if err != nil {
			s.Log.Warn(fmt.Sprintf("Error: failed to parse time %v", strings.TrimSpace(node.FirstChild.Data)))
			continue
		}
		date = time.Date(now.Year(), now.Month(), now.Day(), date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), s.Location)
		ret <- show.Show{
			Time: date,
			Name: strings.TrimSpace(node.FirstChild.NextSibling.NextSibling.Data),
		}
	}
}
