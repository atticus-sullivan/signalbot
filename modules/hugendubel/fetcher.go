package hugendubel

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"log/slog"
)

var (
	ErrNetwork           error = errors.New("Error retreiving from network")
	ErrInvalidResultCode error = errors.New("Error invalid result code")
)

// curl 'https://www.hugendubel.de/wsapi/rest/v1/authentication/anonymousloginjwt' \
//   -H 'authority: www.hugendubel.de' \
//   -H 'accept: */*' \
//   -H 'accept-language: en-US,en;q=0.9,de;q=0.8' \
//   -H 'content-type: application/x-www-form-urlencoded; charset=UTF-8' \
//   -H 'dnt: 1' \
//   -H 'origin: https://www.hugendubel.de' \
//   -H 'referer: https://www.hugendubel.de/de/search/advanced?authors=anthony%20ryan&facets=%3AproductLine_6' \
//   -H 'sec-ch-ua: "Not:A-Brand";v="99", "Chromium";v="112"' \
//   -H 'sec-ch-ua-mobile: ?0' \
//   -H 'sec-ch-ua-platform: "Linux"' \
//   -H 'sec-fetch-dest: empty' \
//   -H 'sec-fetch-mode: cors' \
//   -H 'sec-fetch-site: same-origin' \
//   -H 'user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36' \
//   --data-raw 'username=Hudu-Mobile-Shop-Vollsortiment' \
//   --compressed

type jwtResponse struct {
	Result struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		Version      string `json:"version"`
	} `json:"result"`
	ResultCode int    `json:"resultCode"`
	ResultText string `json:"resultText"`
}

// type hugendubelRequest struct {
// 	Query string `json:"query"` // json encoded
// 	Offset int `json:"offset"`
// 	MaxResults int`json:"maxResults"`
// 	FilterFacets string `json:"filterFacets"`
// 	Ascending bool `json:"ascending"`
// 	SortField string `json:"sortField"`
// }

type hugendubelResponse struct {
	Result struct {
		Articles []struct {
			Active               bool `json:"active"`
			ArticleAttributeView struct {
				AuthorList string `json:"authorList"`
				Title      string `json:"title"`
				Subtitle   string `json:"subtitle"`
			} `json:"articleAttributeView"`
		} `json:"articles"`
		TotalResults int `json:"totalResults"`
	} `json:"result"`
}

type book struct {
	Title    string
	Subtitle string
	Authors  string
}

func (b book) String() string {
	builder := strings.Builder{}

	builder.WriteString(b.Authors)
	builder.WriteString(" -> ")
	builder.WriteString(b.Title)
	if b.Subtitle != "" {
		builder.WriteString(" · ")
		builder.WriteString(b.Subtitle)
	}

	return builder.String()
}

func (b book) AddString() string {
	return "> " + b.String()
}
func (b book) RemString() string {
	return "< " + b.String()
}
func (b book) Equals(o book) bool {
	return b == o
}

type bookItems []book

func (bs bookItems) String() string {
	builder := strings.Builder{}

	first := true
	for _, b := range bs {
		if !first {
			builder.WriteRune('\n')
		} else {
			first = false
		}
		builder.WriteString(b.String())
		builder.WriteRune('\n')
	}

	return builder.String()
}

// fetches stuff. Maybe some day this will have data members (e.g. if caching
// is implemented)
// Has to be instanciated via `NewFetcher`
type Fetcher struct {
	cache *ttlcache.Cache[string, jwtResponse]
	log *slog.Logger
	size uint
}

func NewFetcher(log *slog.Logger, querySize uint) *Fetcher {
	f := &Fetcher{
		cache: ttlcache.New(ttlcache.WithTTL[string, jwtResponse](10*time.Minute), ttlcache.WithDisableTouchOnHit[string, jwtResponse]()),
		log: log,
		size: querySize,
	}
	return f
}

func (f *Fetcher) auth() (*jwtResponse, error) {
	if v := f.cache.Get("all"); v != nil && !v.IsExpired() {
		tmp := v.Value()
		return &tmp, nil
	}

	req, err := http.NewRequest(http.MethodPost, "https://www.hugendubel.de/wsapi/rest/v1/authentication/anonymousloginjwt", bytes.NewBufferString("username=Hudu-Mobile-Shop-Vollsortiment"))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	respS, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer respS.Body.Close()

	if respS.StatusCode != http.StatusOK {
		return nil, ErrNetwork
	}

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, respS.Body)
	if err != nil {
		return nil, err
	}

	resp := jwtResponse{}
	err = json.Unmarshal(buf.Bytes(), &resp)
	if err != nil {
		return nil, err
	}

	if resp.ResultCode != 0 {
		return nil, ErrInvalidResultCode
	}

	f.cache.Set("all", resp, ttlcache.DefaultTTL)
	return &resp, nil
}

func (f *Fetcher) get(qs []query) (bookItems, error) {
	jwt, err := f.auth()
	if err != nil {
		return nil, err
	}

	var ret bookItems

	for _, q := range qs {
		c := make(chan book, 5)
		go f.getStep(q, jwt, int(f.size), c)
		for b := range c {
			ret = append(ret, b)
		}
	}
	f.log.Debug("total amount of items fetched", slog.Int("#", len(ret)))
	return ret, nil
}

type query struct {
	Query  string `yaml:"query"`
	Filter string `yaml:"filter"`
}


func (f *Fetcher) getStep(q query, jwt *jwtResponse, size int, out chan<- book) {
	defer close(out)

	offset := 0
	running := true
	for ; running; offset += size {
		r, err := f.getReader(q, jwt, offset, size)
		if err != nil {
			f.log.Warn(err.Error())
		}

		buf := &bytes.Buffer{}
		_, err = io.Copy(buf, r)
		if err != nil {
			f.log.Warn(err.Error())
		}

		resp := hugendubelResponse{}
		err = json.Unmarshal(buf.Bytes(), &resp)
		if err != nil {
			f.log.Warn(err.Error())
		}

		for _, j := range resp.Result.Articles {
			b := book{
				Title:    j.ArticleAttributeView.Title,
				Subtitle: j.ArticleAttributeView.Subtitle,
				Authors:  j.ArticleAttributeView.AuthorList,
			}
			out <- b
		}

		if len(resp.Result.Articles)+offset >= resp.Result.TotalResults {
			if len(resp.Result.Articles)+offset > resp.Result.TotalResults {
				f.log.Warn("unexpected len vs totalResults", slog.Int("len", len(resp.Result.Articles)+offset), slog.Int("total", resp.Result.TotalResults))
			}
			running = false
		}
		// f.log.Debug("Read items", slog.Int("count", len(resp.Result.Articles)+offset), slog.Any("query", q))
	}
}

// get the content from the internet
func (f *Fetcher) getReader(q query, jwt *jwtResponse, offset int, size int) (io.ReadCloser, error) {
	var err error
	if jwt == nil {
		jwt, err = f.auth()
		if err != nil {
			return nil, err
		}
	}
	j := url.Values{}
	j.Set("query", q.Query)
	j.Set("offset", strconv.Itoa(offset))
	j.Set("maxResults", strconv.Itoa(size))
	j.Set("filterFacets", q.Filter)
	j.Set("ascending", "false")
	j.Set("sortField", "score")

	req, err := http.NewRequest(http.MethodPost, "https://www.hugendubel.de/wsapi/rest/v1/articlesearch/advanced", bytes.NewBufferString(j.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+jwt.Result.AccessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	// req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// j, err := json.Marshal(hugendubelRequest{
// 	Query:        "{\"authors\":[\"anthony ryan\"]}", // TODO how to build this one properly
// 	Offset:       0,
// 	MaxResults:   100, // TODO how would we recognize that this is not enough
// 	FilterFacets: ":productLine_6:language_DE", // TODO enum for productline for languages stuff for building filter string
// 	Ascending:    false,
// 	SortField:    "score", // TODO enum
// })
