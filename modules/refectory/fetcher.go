package refectory

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
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"log/slog"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

var (
	ErrNetwork   error = errors.New("Error retreiving from network")
	ErrParseType error = errors.New("Error parsing high level types")
	ErrParseDesc error = errors.New("Error parsing description")
)

var (
	cascMeal cascadia.Selector = cascadia.MustCompile(".c-menu-dish-list__item")
	cascType cascadia.Selector = cascadia.MustCompile(".stwm-artname")
	cascDesc cascadia.Selector = cascadia.MustCompile(".c-menu-dish__title")
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

// enum with the different Co2 grades
type Co2 rune

// stringer
func (c Co2) String() string {
	return string(c)
}

// enum with the different Water grades
type Water rune

// stringer
func (c Water) String() string {
	return string(c)
}

// represents one meal with name and a list of categories
type Meal struct {
	Name       string
	Categories []Category
	Co2Grade Co2
	WaterGrade Water
}

func (m Meal) Compare(other Meal) int {
	if m.Name != other.Name {
		return strings.Compare(m.Name, other.Name)
	}

	// if m.Co2Grade != other.Co2Grade {
	// 	return int(m.Co2Grade) - int(other.Co2Grade)
	// }
	// if m.WaterGrade != other.WaterGrade {
	// 	return int(m.WaterGrade) - int(other.WaterGrade)
	// }

	for i,c := range m.Categories {
		if !(len(other.Categories) > i) {
			break
		}
		if c != other.Categories[i] {
			return int(c) - int(other.Categories[i])
		}
	}

	return len(m.Categories) - len(other.Categories)
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
			if typeLast != typeStr {
				menu.ordering = append(menu.ordering, typeStr)
			}
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
		// co2
		var co2 Co2
		// water
		var water Water
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
			} else if a.Key == "data-essen-co2-bewertung" {
				if len(a.Val) > 0 {
					co2 = Co2(a.Val[0])
				}
			} else if a.Key == "data-essen-h2o-bewertung" {
				if len(a.Val) > 0 {
					water = Water(a.Val[0])
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
			Co2Grade : co2,
			WaterGrade : water,
		})
	}

	for t,ms := range menu.meals {
		menu.meals[t] = nil
		// slices.SortFunc(ms, func(a Meal, b Meal) int {
		// 	return a.Compare(b)
		// })
		for _, i := range ms {
			if len(menu.meals[t]) == 0 || i.Compare(menu.meals[t][len(menu.meals[t])-1]) != 0 {
				menu.meals[t] = append(menu.meals[t], i)
			}
		}
	}

	return menu, nil
}
