package fernsehserien

import (
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

var (
	ErrNetwork        error = errors.New("Error retreiving from network")
	ErrDateNotFound   error = errors.New("Date not found")
	ErrSenderNotFound error = errors.New("Sender not found")
)

var (
	cascItems  cascadia.Matcher = cascadia.MustCompile("[itemtype=\"http://schema.org/BroadcastEvent\"]:not(.termin-vergangenheit)")
	cascDate   cascadia.Matcher = cascadia.MustCompile("[itemprop=\"startDate\"]")
	cascSender cascadia.Matcher = cascadia.MustCompile("[itemprop=\"name\"]")
)

type sending struct {
	Date   time.Time `yaml:"date"`
	Sender string    `yaml:"sender"`
	Name   string    `yaml:"name"`
}

func (b sending) AddString() string {
	return "> " + b.String()
}
func (b sending) RemString() string {
	return "< " + b.String()
}
func (b sending) Equals(o sending) bool {
	return b == o
}

func (b sending) String() string {
	builder := strings.Builder{}

	builder.WriteString(b.Date.Format("2006-01-02 15:04"))
	builder.WriteString(": ")
	builder.WriteString(b.Name)
	builder.WriteString(" -> ")
	builder.WriteString(b.Sender)

	return builder.String()
}

type sendings []sending

func (b sendings) String() string {
	builder := strings.Builder{}

	first := true
	for _, i := range b {
		if !first {
			builder.WriteRune('\n')
		} else {
			first = false
		}
		builder.WriteString(i.String())
		builder.WriteRune('\n')
	}

	return builder.String()
}

// fetches stuff. Maybe some day this will have data members (e.g. if caching
// is implemented)
type Fetcher struct{}

// parse the content from an arbitrary reader (can be a file, a network
// response body or something else)
func (f *Fetcher) getFromReaders(readers map[string]io.ReadCloser, unavailableSenders map[string]bool) (sendings, error) {
	var ret sendings
	var ret_old sendings

	for name, reader := range readers {
		root, err := html.Parse(reader)
		if err != nil {
			return nil, err
		}

		items := cascadia.QueryAll(root, cascItems)
		// expand the ret slice by the length of the items slice
		ret, ret_old = make(sendings, len(ret), len(ret)+len(items)), ret
		copy(ret, ret_old)
		for _, i := range items {

			dateN := cascadia.Query(i, cascDate)
			var date time.Time
			for _, attr := range dateN.Attr {
				if attr.Key == "datetime" {
					date, err = time.Parse(time.RFC3339, attr.Val)
					if err != nil {
						return nil, err
					}
				}
			}
			if date.IsZero() {
				return nil, ErrDateNotFound
			}

			senderN := cascadia.Query(i, cascSender)
			var sender string
			for _, attr := range senderN.Attr {
				if attr.Key == "content" {
					sender = attr.Val
				}
			}
			if sender == "" {
				return nil, ErrSenderNotFound
			}
			if val, unavailable := unavailableSenders[sender]; unavailable && val {
				continue
			}

			it := sending{
				Name:   name,
				Date:   date,
				Sender: sender,
			}
			ret = append(ret, it)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Date.Before(ret[j].Date)
	})
	return ret, nil
}

// get the content from the internet
func (f *Fetcher) getReaders(urls map[string]string) (map[string]io.ReadCloser, error) {
	ret := make(map[string]io.ReadCloser)
	for name, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, ErrNetwork
		}

		ret[name] = resp.Body
	}
	return ret, nil
}
