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
	"io"
	"log/slog"
	"strings"
	"time"
)

var (
	ErrNetwork   error = errors.New("Error retreiving from network")
	ErrParseType error = errors.New("Error parsing high level types")
	ErrParseDesc error = errors.New("Error parsing description")
	ErrNotOpenThatDay error = errors.New("Refectory not open that day")
	ErrDateNotFound error = errors.New("Date field was not found")
	ErrDateUnparsable error = errors.New("Date field was unparsable")
)

type FetcherReadCloser interface {
	io.ReadCloser
}

type FetcherInter[T FetcherReadCloser] interface {
	init(log *slog.Logger)
	getFromReader(reader T) (Menu, error)
	getReader(mensa_id uint, date time.Time) (T, error)
}

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

