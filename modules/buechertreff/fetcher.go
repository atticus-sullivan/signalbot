package buechertreff

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

type bookItem struct {
	pos      string
	name     string
	origName string
	pub      string
}

func (b bookItem) String() string {
	builder := strings.Builder{}

	builder.WriteString(b.pos)
	builder.WriteRune('\n')
	builder.WriteString(b.name)
	builder.WriteRune('\n')
	if b.origName != "" {
		builder.WriteString(b.origName)
		builder.WriteRune(' ')
	}
	builder.WriteString(b.pub)

	return builder.String()
}

type bookItems []bookItem

func (b bookItems) String() string {
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

type Fetcher struct {
}

func (f *Fetcher) get(url string) (bookItems, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("")
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	items := cascadia.QueryAll(root, cascadia.MustCompile(".containerHeadline"))
	bookItems := make(bookItems, 0, len(items))
	for _, i := range items {
		it := bookItem{}

		posN := cascadia.Query(i, cascadia.MustCompile("[itemprop=\"position\"]"))
		if posN == nil || posN.FirstChild == nil {
			return nil, fmt.Errorf("")
		}
		it.pos = strings.TrimSpace(posN.FirstChild.Data)

		nameN := cascadia.Query(i, cascadia.MustCompile("[itemprop=\"name\"]"))
		if nameN == nil || nameN.FirstChild == nil {
			return nil, fmt.Errorf("")
		}
		it.name = strings.TrimSpace(nameN.FirstChild.Data)

		pubN := cascadia.Query(i, cascadia.MustCompile("[itemprop=\"datePublished\"]"))
		if pubN == nil || pubN.FirstChild == nil {
			return nil, fmt.Errorf("")
		}
		it.pub = strings.TrimSpace(pubN.FirstChild.Data)

		origNameN := pubN.PrevSibling
		if origNameN != nil && origNameN.Type == html.TextNode {
			it.origName = strings.TrimSpace(origNameN.Data)
		}
		bookItems = append(bookItems, it)
	}

	return bookItems, nil
}
