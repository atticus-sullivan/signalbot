package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func mustLoadLocation(l string) *time.Location {
	r, err := time.LoadLocation(l)
	if err != nil {
		panic(err)
	}
	return r
}

var loc *time.Location = mustLoadLocation("Europe/Berlin")

// TODO maybe make this type anonymous -> use struct{} in entry struct instead
// of contentLine
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

func (e *entry) String() string {
	if len(e.Content) <= 0 {
		return ""
	}
	builder := strings.Builder{}

	builder.WriteString("🔵 ")
	builder.WriteString(e.Topline)
	builder.WriteRune('\n')
	// strip html tags
	builder.WriteString(regexp.MustCompile("<.*?>").ReplaceAllLiteralString(e.Content[0].Value, ""))
	builder.WriteRune('\n')
	builder.WriteString(e.Webpage)

	return builder.String()
}

type entries []entry

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
	Headline string `json:"headline" yaml:"headline"`
	Text     string `json:"text" yaml:"text"`
	Url      string `json:"url" yaml:"url"`
	LinkText string `json:"linkText" yaml:"linkText"`

	DateS string    `json:"date"`
	Date  time.Time `yaml:"date"`
}

func (e *breakingResp) String() string {
	builder := strings.Builder{}

	builder.WriteString("⚡️")
	builder.WriteString(e.LinkText)
	builder.WriteString(": ")
	builder.WriteString(e.Headline)
	builder.WriteRune('\n')
	builder.WriteRune('\n')
	builder.WriteString(e.Text)
	builder.WriteRune('\n')
	builder.WriteRune('\n')
	builder.WriteString(e.Url)

	return builder.String()
}
func (e breakingResp) AddString() string {
	return e.String()
}

// removal should not be displayed
func (e breakingResp) RemString() string {
	return ""
}

func (e *breakingResp) IsZero() bool {
	return e.Text == ""
}

type breakings []breakingResp

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

type Fetcher struct {
}

func (f *Fetcher) getNews() (entries, error) {
	response, err := http.Get("https://www.tagesschau.de/api2/homepage/")
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching failed with status: %s", response.Status)
	}
	defer response.Body.Close()

	var resp homepageResp
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
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

func (f *Fetcher) getBreaking() (breakings, error) {
	response, err := http.Get("https://www.tagesschau.de/ipa/v1/web/headerapp/")
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching failed with status: %s", response.Status)
	}
	defer response.Body.Close()

	var resp breakingResp
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, err
	}

	// resp = breakingResp{
	// 	Headline: "Merkel kündigt Ende der Impf-Priorisierung im Juni an",
	// 	Text:     "Nach dem Impfgipfel von Bund und Ländern hat Kanzlerin Merkel den Plan bekräftigt, dass die Priorisierung beim Impfen gegen das Coronavirus im Juni aufgehoben werden kann. Dann sei noch nicht jeder geimpft - aber es gebe die Möglichkeit, einen Termin zu erhalten. Zu möglichen Lockerungen für Geimpfte und Genesene fiel bei dem Gipfel keine Entscheidung.",
	// 	Url:      "https://www.tagesschau.de/eilmeldung/eilmeldung-5565.html",
	// 	LinkText: "Eilmeldung",
	// 	DateS:    "26.04.2021 - 18:02 Uhr",
	// }

	if resp.DateS != "" {
		// date comes as "26.04.2021 - 18:02 Uhr"
		resp.Date, err = time.ParseInLocation("02.01.2006 - 15:04 Uhr", resp.DateS, loc)
		if err != nil {
			return nil, err
		}
	}

	if resp.IsZero() {
		return breakings{}, nil
	} else {
		return breakings{resp}, nil
	}
}
