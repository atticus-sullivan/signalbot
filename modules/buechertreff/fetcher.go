package buechertreff

// signalbot
// Copyright (C) 2024  Lukas Heindl
// 
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// 
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

var (
	ErrNetwork error = errors.New("Error retreiving from network")

	ErrParsePos  error = errors.New("Error parsing position")
	ErrParseName error = errors.New("Error parsing name")
	ErrParsePub  error = errors.New("Error parsing publication year")
)

var (
	cascItems cascadia.Matcher = cascadia.MustCompile(".containerHeadline")
	cascPos   cascadia.Matcher = cascadia.MustCompile("[itemprop=\"position\"]")
	cascName  cascadia.Matcher = cascadia.MustCompile("[itemprop=\"name\"]")
	cascPub   cascadia.Matcher = cascadia.MustCompile("[itemprop=\"datePublished\"]")
)

// represents one book
type bookItem struct {
	pos      string
	name     string
	origName string
	pub      string
}

// implements stringer
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

// multiple books (new type so that this can implement stringer new)
type bookItems []bookItem

// implements stringer
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

// fetches stuff. Maybe some day this will have data members (e.g. if caching
// is implemented)
type Fetcher struct{}

// parse the content from an arbitrary reader (can be a file, a network
// response body or something else)
func (f *Fetcher) getFromReader(r io.Reader) (bookItems, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	items := cascadia.QueryAll(root, cascItems)
	bookItems := make(bookItems, 0, len(items))
	for _, i := range items {
		it := bookItem{}

		posN := cascadia.Query(i, cascPos)
		if posN == nil || posN.FirstChild == nil {
			return nil, ErrParsePos
		}
		it.pos = strings.TrimSpace(posN.FirstChild.Data)

		nameN := cascadia.Query(i, cascName)
		if nameN == nil || nameN.FirstChild == nil {
			return nil, ErrParseName
		}
		it.name = strings.TrimSpace(nameN.FirstChild.Data)

		pubN := cascadia.Query(i, cascPub)
		if pubN == nil || pubN.FirstChild == nil {
			return nil, ErrParsePub
		}
		it.pub = strings.TrimSpace(pubN.FirstChild.Data)

		// allowed to be missing
		origNameN := pubN.PrevSibling
		if origNameN != nil && origNameN.Type == html.TextNode {
			it.origName = strings.TrimSpace(origNameN.Data)
		}
		bookItems = append(bookItems, it)
	}

	return bookItems, nil
}

// get the content from the internet
func (f *Fetcher) getReader(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, ErrNetwork
	}

	return resp.Body, nil
}
