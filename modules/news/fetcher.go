package news

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	ErrNetwork error = errors.New("Error retreiving from network")
)

var (
	reHtmlTags *regexp.Regexp = regexp.MustCompile("<.*?>")
)

func mustLoadLocation(l string) *time.Location {
	r, err := time.LoadLocation(l)
	if err != nil {
		panic(err)
	}
	return r
}

var loc *time.Location = mustLoadLocation("Europe/Berlin")

type contentLine struct {
	Value string `json:"value" yaml:"value"`
	Type  string `json:"type" yaml:"type"`
}

type entry struct {
	Date    time.Time `json:"date" yaml:"date"`
	Webpage string    `json:"detailsWeb" yaml:"webpage"`

	Title         string        `json:"title" yaml:"title"`
	Topline       string        `json:"topline" yaml:"topline"`
	FirstSentence string        `json:"firstSentence" yaml:"firstSentence"`
	Content       []contentLine `json:"content" yaml:"content"`
}

// stringer
func (e *entry) String() string {
	if len(e.Content) <= 0 {
		return ""
	}
	builder := strings.Builder{}

	builder.WriteString("🔵 ")
	builder.WriteString(e.Topline)
	builder.WriteRune('\n')
	// strip html tags
	builder.WriteString(reHtmlTags.ReplaceAllLiteralString(e.Content[0].Value, ""))
	builder.WriteRune('\n')
	builder.WriteString(e.Webpage)

	return builder.String()
}

// new type so that it can implement stringer
type entries []entry

// stringer
func (e *entries) String() string {
	builder := strings.Builder{}
	first := true
	for _, ei := range *e {
		str := ei.String()
		if str == "" {
			continue
		}
		if !first {
			builder.WriteString("\n\n")
		} else {
			first = false
		}
		builder.WriteString(str)
	}
	return builder.String()
}

type homepageResp struct {
	News entries `json:"news" yaml:"news"`
}

type breakingResp struct {
	BreakingNews breakings `yaml:"breakingNews"`
}

type breaking struct {
	Headline string `json:"headline" yaml:"headline"`
	Text     string `json:"text" yaml:"text"`
	Url      string `json:"url" yaml:"url"`
	LinkText string `json:"linkText" yaml:"linkText"`
	Id       string `json:"id" yaml:"id"`

	DateS string    `json:"date"`
	Date  time.Time `yaml:"date"`
}

// stringer
func (e *breaking) String() string {
	builder := strings.Builder{}

	builder.WriteString("⚡️")
	builder.WriteString(e.LinkText)
	builder.WriteString(": ")
	builder.WriteString(e.Headline)
	builder.WriteRune('\n')
	builder.WriteRune('\n')
	builder.WriteString(e.Text)
	builder.WriteRune('\n')
	builder.WriteString(e.Url)

	return builder.String()
}

// differ
func (e breaking) AddString() string {
	return e.String()
}

// removal should not be displayed
// differ
func (e breaking) RemString() string {
	return ""
}

func (e breaking) Equals(o breaking) bool {
	return e.Id == o.Id
}

// new type so that it can implement stringer
type breakings []breaking

// stringer
func (e *breakings) String() string {
	builder := strings.Builder{}
	first := true
	for _, v := range *e {
		if !first {
			builder.WriteRune('\n')
		} else {
			first = false
		}
		builder.WriteString(v.String())
	}
	return builder.String()
}

// fetches stuff. Maybe some day this will have data members (e.g. if caching
// is implemented)
type Fetcher struct{}

// get the content from the internet
func (f *Fetcher) getNewsReader() (io.ReadCloser, error) {
	response, err := http.Get("https://www.tagesschau.de/api2/homepage/")
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, ErrNetwork
	}
	return response.Body, nil
}

// parse the content from an arbitrary reader (can be a file, a network
// response body or something else)
func (f *Fetcher) getNewsFromReader(reader io.Reader) (entries, error) {
	var resp homepageResp
	if err := json.NewDecoder(reader).Decode(&resp); err != nil {
		return nil, err
	}
	// remove weather news if present
	if len(resp.News) > 0 && resp.News[len(resp.News)-1].Topline == "Vorhersage" {
		resp.News = resp.News[:len(resp.News)-1]
	}
	if len(resp.News) > 0 && resp.News[len(resp.News)-1].Topline == "Sportschau" {
		resp.News = resp.News[:len(resp.News)-1]
	}
	return resp.News, nil
}

// get the content from the internet
func (f *Fetcher) getBreakingReader() (io.ReadCloser, error) {
	response, err := http.Get("https://www.tagesschau.de/json/headerapp")
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, ErrNetwork
	}
	return response.Body, nil
}

// parse the content from an arbitrary reader (can be a file, a network
// response body or something else)
func (f *Fetcher) getBreakingFromReader(reader io.ReadCloser) (breakings, error) {
	var respB breakingResp
	buf := &bytes.Buffer{}
	if _,err := io.Copy(buf, reader); err != nil {
		return nil, err
	}
	var err error
	if err := json.Unmarshal(buf.Bytes(), &respB); err != nil {
		// TODO only temporary for debugging
		ferr, _ := os.CreateTemp("", time.Now().Format("2006-01-02_15-04-05_signalbotBreaking.json"))
		ferr.Write(buf.Bytes())
		return nil, err
	}
	resp := respB.BreakingNews

	// resp = breakingResp{
	// 	Headline: "Merkel kündigt Ende der Impf-Priorisierung im Juni an",
	// 	Text:     "Nach dem Impfgipfel von Bund und Ländern hat Kanzlerin Merkel den Plan bekräftigt, dass die Priorisierung beim Impfen gegen das Coronavirus im Juni aufgehoben werden kann. Dann sei noch nicht jeder geimpft - aber es gebe die Möglichkeit, einen Termin zu erhalten. Zu möglichen Lockerungen für Geimpfte und Genesene fiel bei dem Gipfel keine Entscheidung.",
	// 	Url:      "https://www.tagesschau.de/eilmeldung/eilmeldung-5565.html",
	// 	LinkText: "Eilmeldung",
	// 	DateS:    "26.04.2021 - 18:02 Uhr",
	// }

	for i,r := range resp {
		if r.DateS != "" {
			// date comes as "Stand: 27.04.2023 12:53 Uhr"
			resp[i].Date, err = time.ParseInLocation("Stand: 02.01.2006 15:04 Uhr", r.DateS, loc)
			if err != nil {
				return nil, err
			}
		}
	}

	return resp, nil
}
