package fernsehserien

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

type sending struct {
	Date time.Time `yaml:"date"`
	Sender string `yaml:"sender"`
	Name string `yaml:"name"`
}

func (b sending) AddString() string {
	return "> " + b.String()
}
func (b sending) RemString() string {
	return "< " + b.String()
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
	for _,i := range b {
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

func Get(urls map[string]string, unavailableSenders map[string]bool) (sendings, error) {
	var ret sendings
	var ret_old sendings
	for name, url := range urls {
		resp,err := http.Get(url)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("")
		}

		root,err := html.Parse(resp.Body)
		if err != nil {
			return nil, err
		}

		items := cascadia.QueryAll(root, cascadia.MustCompile("[itemtype=\"http://schema.org/BroadcastEvent\"]:not(.termin-vergangenheit)"))
		ret, ret_old = make(sendings, len(ret), len(ret)+len(items)), ret
		copy(ret, ret_old)
		for _,i := range items {

			dateN := cascadia.Query(i, cascadia.MustCompile("[itemprop=\"startDate\"]"))
			var date time.Time
			for _,attr := range dateN.Attr {
				if attr.Key == "datetime" {
					date, err = time.Parse(time.RFC3339, attr.Val)
					if err != nil {
						return nil, err
					}
				}
			}
			if date.IsZero() {
				return nil, fmt.Errorf("")
			}

			senderN := cascadia.Query(i, cascadia.MustCompile("[itemprop=\"name\"]"))
			var sender string
			for _,attr := range senderN.Attr {
				if attr.Key == "content" {
					sender = attr.Val
				}
			}
			if sender == "" {
				return nil, fmt.Errorf("")
			}
			if val,unavailable := unavailableSenders[sender]; unavailable && val {
				continue
			}

			it := sending{
				Name: name,
				Date: date,
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
