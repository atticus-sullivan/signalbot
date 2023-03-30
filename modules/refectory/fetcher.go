package refectory

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/andybalholm/cascadia"
	"golang.org/x/exp/slog"
	"golang.org/x/net/html"
)

type Category rune

const (
	BEEF  Category = '🐄'
	PORK  Category = '🐷'
	VEGGY Category = '🥕'
	VEGAN Category = '🥑'
	FISH  Category = '🐟'
)

func (c Category) String() string {
	return string(c)
}

type Meal struct {
	Name       string
	Categories []Category
}

func (m Meal) String() string {
	builder := strings.Builder{}

	builder.WriteString(m.Name)
	builder.WriteRune(' ')
	for _, c := range m.Categories {
		builder.WriteString(c.String())
	}

	return builder.String()
}

type Menu struct {
	meals    map[string][]Meal
	ordering []string
}

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

// could implement caching if necessary
type Fetcher struct {
	log *slog.Logger
}

func (f *Fetcher) getMenuString(mensa string, mensa_id uint, date time.Time) (string, error) {
	menu, err := f.getMenu(mensa_id, date)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s on %s\n", mensa, date.Format("2006-01-02")) + menu.String(), nil
}

func (f *Fetcher) getMenu(mensa_id uint, date time.Time) (Menu, error) {
	root, err := f.download(mensa_id, date)
	if err != nil {
		return Menu{}, fmt.Errorf("Failed downloading menu for %s: %v", date.Format("2006-01-02"), err)
	}
	menu, err := f.parse(root)
	if err != nil {
		return menu, fmt.Errorf("Failed parsing menu for %s: %v", date.Format("2006-01-02"), err)
	}
	return menu, nil
}

var MEAL_URL_TEMPLATE string = "https://www.studentenwerk-muenchen.de/mensa/speiseplan/speiseplan_%s_%d_-de.html"

var NotOpenThatDay error = errors.New("Refectory not open that day")

func (f *Fetcher) download(mensa_id uint, date time.Time) (*html.Node, error) {
	url := fmt.Sprintf(MEAL_URL_TEMPLATE, date.Format("2006-01-02"), mensa_id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, NotOpenThatDay
	case http.StatusOK:
	default:
		return nil, fmt.Errorf("Response: %s", resp.Status)
	}

	r, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	return r, nil
}

var (
	selMeal cascadia.Selector = cascadia.MustCompile(".c-schedule__list-item")
	selType cascadia.Selector = cascadia.MustCompile(".stwm-artname")
	selDesc cascadia.Selector = cascadia.MustCompile(".js-schedule-dish-description")
)

func (f *Fetcher) parse(root *html.Node) (Menu, error) {
	menu := Menu{
		meals:    make(map[string][]Meal),
		ordering: make([]string, 0),
	}
	// Load the HTML document
	typeLast := ""
	for _, meal := range cascadia.QueryAll(root, selMeal) {
		// type
		var typeStr string
		typeNodes := cascadia.QueryAll(meal, selType)
		if len(typeNodes) != 1 {
			return Menu{}, fmt.Errorf("invalid amount (%d) of typeNodes found", len(typeNodes))
		}
		if typeNodes[0].FirstChild != nil {
			typeStr = typeNodes[0].FirstChild.Data
			menu.ordering = append(menu.ordering, typeStr)
			typeLast = typeStr
		} else {
			typeStr = typeLast
		}
		// fmt.Println(typeStr)

		// description
		var descStr string
		descNodes := cascadia.QueryAll(meal, selDesc)
		if len(descNodes) != 1 {
			return Menu{}, fmt.Errorf("invalid amount (%d) of descNodes found", len(descNodes))
		}
		for n := descNodes[0].FirstChild; n != nil; n = n.NextSibling {
			if n.Type == html.TextNode {
				descStr = n.Data
				break
			}
		}
		// fmt.Println(descStr)

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
		// fmt.Println(cats)
		_, ok := menu.meals[typeStr]
		if !ok {
			menu.meals[typeStr] = make([]Meal, 0)
		}
		menu.meals[typeStr] = append(menu.meals[typeStr], Meal{
			Name:       descStr,
			Categories: cats,
		})
	}
	return menu, nil
}
