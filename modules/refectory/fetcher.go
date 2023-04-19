package refectory

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/andybalholm/cascadia"
	"golang.org/x/exp/slog"
	"golang.org/x/net/html"
)

var (
	ErrNetwork   error = errors.New("Error retreiving from network")
	ErrParseType error = errors.New("Error parsing high level types")
	ErrParseDesc error = errors.New("Error parsing description")
)

var (
	cascMeal cascadia.Selector = cascadia.MustCompile(".c-schedule__list-item")
	cascType cascadia.Selector = cascadia.MustCompile(".stwm-artname")
	cascDesc cascadia.Selector = cascadia.MustCompile(".js-schedule-dish-description")
)

// enum with the different food categories
type Category rune

const (
	BEEF  Category = '🐄'
	PORK  Category = '🐷'
	VEGGY Category = '🥕'
	VEGAN Category = '🥑'
	FISH  Category = '🐟'
)

// stringer
func (c Category) String() string {
	return string(c)
}

// represents one meal with name and a list of categories
type Meal struct {
	Name       string
	Categories []Category
}

// stringer
func (m Meal) String() string {
	builder := strings.Builder{}

	builder.WriteString(m.Name)
	builder.WriteRune(' ')
	for _, c := range m.Categories {
		builder.WriteString(c.String())
	}

	return builder.String()
}

// a menu is an enumeration of all available meals
type Menu struct {
	meals    map[string][]Meal
	ordering []string
}

// stringer
func (m Menu) String() string {
	builder := strings.Builder{}

	for _, t := range m.ordering {
		ms, ok := m.meals[t]
		if !ok {
			continue
		}
		builder.WriteRune('*')
		builder.WriteString(t)
		builder.WriteRune('*')
		builder.WriteRune(':')
		builder.WriteRune('\n')
		for _, meal := range ms {
			builder.WriteString(meal.String())
			builder.WriteRune('\n')
		}
	}
	builder.WriteString(VEGAN.String())
	builder.WriteString(" = vegan, ")
	builder.WriteString(VEGGY.String())
	builder.WriteString(" = vegetarisch\n")
	builder.WriteString(PORK.String())
	builder.WriteString(" = Schwein, ")
	builder.WriteString(BEEF.String())
	builder.WriteString(" = Rind, ")
	builder.WriteString(FISH.String())
	builder.WriteString(" = Fisch")

	return builder.String()
}

// fetches stuff. (e.g. if caching can be implemented at this level)
type Fetcher struct {
	log *slog.Logger
}

var MEAL_URL_TEMPLATE string = "https://www.studentenwerk-muenchen.de/mensa/speiseplan/speiseplan_%s_%d_-de.html"

var ErrNotOpenThatDay error = errors.New("Refectory not open that day")

// get the content from the internet
func (f *Fetcher) getReader(mensa_id uint, date time.Time) (io.ReadCloser, error) {
	url := fmt.Sprintf(MEAL_URL_TEMPLATE, date.Format("2006-01-02"), mensa_id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, ErrNotOpenThatDay
	case http.StatusOK:
	default:
		return nil, ErrNetwork
	}

	return resp.Body, nil
}

// parse the content from an arbitrary reader (can be a file, a network
// response body or something else)
func (f *Fetcher) getFromReader(reader io.Reader) (Menu, error) {
	menu := Menu{
		meals:    make(map[string][]Meal),
		ordering: make([]string, 0),
	}

	root, err := html.Parse(reader)
	if err != nil {
		return menu, err
	}

	// Load the HTML document
	typeLast := ""
	for _, meal := range cascadia.QueryAll(root, cascMeal) {
		// type
		var typeStr string
		typeNodes := cascadia.QueryAll(meal, cascType)
		if len(typeNodes) != 1 {
			return Menu{}, ErrParseType
		}
		if typeNodes[0].FirstChild != nil {
			typeStr = typeNodes[0].FirstChild.Data
			menu.ordering = append(menu.ordering, typeStr)
			typeLast = typeStr
		} else {
			typeStr = typeLast
		}

		// description
		var descStr string
		descNodes := cascadia.QueryAll(meal, cascDesc)
		if len(descNodes) != 1 {
			return Menu{}, ErrParseDesc
		}
		for n := descNodes[0].FirstChild; n != nil; n = n.NextSibling {
			if n.Type == html.TextNode {
				descStr = n.Data
				break
			}
		}

		// categories
		cats := make([]Category, 0)
		for _, a := range meal.Attr {
			if a.Key == "data-essen-fleischlos" {
				switch a.Val {
				case "0":
				case "1":
					cats = append(cats, VEGGY)
				case "2":
					cats = append(cats, VEGAN)
				default:
					f.log.Warn(fmt.Sprintf("unknown 'data-essen-fleischlos': %s", a.Val))
				}
			} else if a.Key == "data-essen-typ" {
				for _, val := range strings.Split(a.Val, ",") {
					switch val {
					case "":
					case "R":
						cats = append(cats, BEEF)
					case "S":
						cats = append(cats, PORK)
					default:
						f.log.Warn(fmt.Sprintf("unknown 'data-essen-typ': %s", val))
					}
				}
			} else if a.Key == "data-essen-allergene" {
				for _, val := range strings.Split(a.Val, ",") {
					switch val {
					case "":
					case "Fi":
						cats = append(cats, FISH)
					default:
					}
				}
			}
		}
		_, ok := menu.meals[typeStr]
		if !ok {
			menu.meals[typeStr] = make([]Meal, 0)
		}
		menu.meals[typeStr] = append(menu.meals[typeStr], Meal{
			Name:       strings.TrimSpace(descStr),
			Categories: cats,
		})
	}
	return menu, nil
}
