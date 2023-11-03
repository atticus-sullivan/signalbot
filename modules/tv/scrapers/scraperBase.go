package scrapers

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
