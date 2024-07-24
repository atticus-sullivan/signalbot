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
	"io"
	"os"
	"testing"

	"log/slog"
)

func nopLog() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestFetcher(t *testing.T) {
	log := nopLog()
	ref, err := NewRefectory(log, "./")
	if err != nil {
		panic(err)
	}

	f, err := os.Open("test1.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	menu, err := ref.fetcher.getFromReader(f)
	if err != nil {
		panic(err)
	}

	ord := []string{"Pasta", "Pizza", "Grill", "Wok", "Studitopf", "Fleisch", "Vegan", "Beilagen"}
	meals := map[string][]Meal{
		"Pasta": {
			Meal{
				Name:       "Pasta mit Sojabolognese",
				Categories: []Category{VEGAN},
				Co2Grade: Co2('A'),
				WaterGrade: Water('B'),
			}},
		"Pizza": {
			Meal{
				Name:       "Pizza Margherita mit Mozzarella",
				Categories: []Category{VEGGY},
				Co2Grade: Co2('B'),
				WaterGrade: Water('B'),
			}},
		"Grill": {
			Meal{
				Name:       "Bierbrauersteak (1 Stück) (S vom Strohschwein) mit Zwiebelschmelze",
				Categories: []Category{},
				Co2Grade: Co2('C'),
				WaterGrade: Water('A'),
			}},
		"Wok": {
			Meal{
				Name:       "Puten-Gemüse-Curry",
				Categories: []Category{},
				Co2Grade: Co2('B'),
				WaterGrade: Water('B'),
			}},
		"Studitopf": {
			Meal{
				Name:       "Asiatisches Gemüse mit Chinakohl (scharf)",
				Categories: []Category{VEGAN},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
			Meal{Name: "Tomatenrahmsuppe",
				Categories: []Category{},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			}},
		"Fleisch": {
			Meal{
				Name:       "Fleischpflanzerl mit Kümmelsauce",
				Categories: []Category{BEEF, PORK},
				Co2Grade: Co2('B'),
				WaterGrade: Water('B'),
			}},
		"Vegan": {
			Meal{
				Name:       "Ofengemüse mit weißem Bohnenpüree und Basilikumpesto",
				Categories: []Category{VEGAN},
				Co2Grade: Co2('A'),
				WaterGrade: Water('C'),
			}},
		"Beilagen": {
			Meal{
				Name:       "Asia Reis Bowl mit Tofu",
				Categories: []Category{VEGAN},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
			Meal{Name: "Basmatireis",
				Categories: []Category{VEGAN},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
			Meal{Name: "Petersilienkartoffeln",
				Categories: []Category{VEGAN},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
			Meal{Name: "Täglich frisches Gemüse",
				Categories: []Category{},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
			Meal{Name: "Mediterranes Gemüse",
				Categories: []Category{VEGAN},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
			Meal{Name: "Fenchel-Tomaten-Gemüse",
				Categories: []Category{VEGAN},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
			Meal{Name: "Täglich frische Dessertbar",
				Categories: []Category{},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
			Meal{Name: "Täglich frisches Salatbuffet",
				Categories: []Category{},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
			Meal{Name: "Mousse au chocolat mit Orangen",
				Categories: []Category{},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
			Meal{Name: "Frischer Obstsalat",
				Categories: []Category{VEGAN},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
			Meal{Name: "Frische Melone mit Minze",
				Categories: []Category{VEGAN},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
			Meal{Name: "Pina Colada: Smoothie mit Ananas und Kokos",
				Categories: []Category{VEGAN},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
			Meal{Name: "Grüner Apfel-Zucchini-Saft mit Minze",
				Categories: []Category{VEGAN},
				Co2Grade: Co2('B'),
				WaterGrade: Water('A'),
			},
		},
	}
	if len(menu.meals) != len(ord) {
		t.Fatalf("Invalid number of meal buckets parsed (was %d, should %d)", len(menu.meals), len(ord))
	}
	if len(menu.ordering) != len(ord) {
		t.Fatalf("Invalid number of meal buckets parsed  (was %d, should %d)", len(menu.ordering), len(ord))
	}
	for i := range ord {
		if menu.ordering[i] != ord[i] {
			t.Fatalf("Wrong ordering")
		}
		if len(menu.meals[ord[i]]) != len(meals[ord[i]]) {
			t.Fatalf("Wrong meals for %s", ord[i])
		}
		for j := range meals[ord[i]] {
			m := menu.meals[ord[i]][j]
			m_ref := meals[ord[i]][j]
			if m.Name != m_ref.Name {
				t.Fatalf("Wrong meal name. %s (should: %s)", m.Name, m_ref.Name)
			}
			if len(m.Categories) != len(m_ref.Categories) {
				t.Fatalf("Wrong amount of categories (%s: was %v, should %v)", m.Name, m.Categories, m_ref.Categories)
			}
			if m.Co2Grade != m_ref.Co2Grade {
				t.Fatalf("Wrong co2 grade (%s: was %v, should %v)", m.Name, m.Co2Grade, m_ref.Co2Grade)
			}
			if m.WaterGrade != m_ref.WaterGrade {
				t.Fatalf("Wrong water grade (%s: was %v, should %v)", m.Name, m.WaterGrade, m_ref.WaterGrade)
			}
			for k := range m_ref.Categories {
				if m.Categories[k] != m_ref.Categories[k] {
					t.Fatalf("Wrong category")
				}
			}
		}
	}

	// fo,_ := os.Create("test1.out")
	// fo.Write([]byte(menu.String()))
	// fo.Close()

	str := menu.String()
	out, err := os.ReadFile("test1.out")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("test1.tout", []byte(str), 0666)
	if err != nil {
		panic(err)
	}

	if str != string(out) {
		t.Fatalf("formatting is wrong")
	}
}
