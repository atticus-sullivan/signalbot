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

	items := cascadia.QueryAll(root, cascadia.MustCompile("epg-broadcast-row"))
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
		for _, t := range cascadia.QueryAll(i, cascadia.MustCompile(".teaser-title")) {
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
			Time: date,
			Name: strings.TrimSpace(name.String()),
		}
		ret <- retS
	}
}
