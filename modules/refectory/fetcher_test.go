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
	return slog.New(slog.HandlerOptions{Level: slog.LevelError}.NewTextHandler(io.Discard))
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

	ord := []string{"Pasta", "Grill", "Wok", "Studitopf", "Vegetarisch/fleischlos", "Fleisch", "Beilagen"}
	meals := map[string][]Meal{
		"Pasta": {Meal{
			Name:       "Tortellini mit Ricotta-Spinat-Füllung in Kräutersauce",
			Categories: []Category{VEGGY},
		}},
		"Grill": {Meal{
			Name:       "Cevapcici mit Ajvar",
			Categories: []Category{BEEF},
		}},
		"Wok": {Meal{
			Name:       "Tofugeschnetzeltes China-Town",
			Categories: []Category{VEGAN},
		}},
		"Studitopf": {Meal{
			Name:       "Bulgur indische Art",
			Categories: []Category{VEGAN},
		},
			Meal{Name: "Tagessuppe",
				Categories: []Category{},
			}},
		"Vegetarisch/fleischlos": {Meal{
			Name:       "Hausgemachte Gemüsequiche mit Schnittlauchdip",
			Categories: []Category{VEGGY},
		}},
		"Fleisch": {Meal{
			Name:       "Balinesisches Kokoshähnchen",
			Categories: []Category{},
		}},
		"Beilagen": {Meal{
			Name:       "Countrykartoffeln",
			Categories: []Category{VEGGY},
		},
			Meal{Name: "Erbsen, natur",
				Categories: []Category{VEGAN},
			},
			Meal{Name: "Djuvecreis",
				Categories: []Category{VEGAN},
			},
			Meal{Name: "Täglich frisches Gemüse",
				Categories: []Category{},
			},
			Meal{Name: "Zucchinis mit Karotten,Schwarzwurzel und grünen Bohnen",
				Categories: []Category{VEGAN},
			},
			Meal{Name: "Täglich frische Dessertbar",
				Categories: []Category{},
			},
			Meal{Name: "Täglich frische Salatbar",
				Categories: []Category{},
			},
			Meal{Name: "Vanillecreme mit passierten Himbeeren",
				Categories: []Category{VEGGY},
			},
			Meal{Name: "Frische Melone mit Minze",
				Categories: []Category{VEGAN},
			},
			Meal{Name: "Sunset: Himbeer-Bananen-Smoothie mit Orangensaft",
				Categories: []Category{VEGAN},
			},
			Meal{Name: "Pfirsich-Mango-Orangen-Saft",
				Categories: []Category{VEGAN},
			},
		},
	}
	if len(menu.meals) != len(ord) {
		t.Fatalf("Invalid number of meal buckets parsed")
	}
	if len(menu.ordering) != len(ord) {
		t.Fatalf("Invalid number of meal buckets parsed")
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
				t.Fatalf("Wrong amount of categories")
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

	if str != string(out) {
		t.Fatalf("formatting is wrong")
	}
}
