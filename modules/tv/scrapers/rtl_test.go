package scrapers_test

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
	"os"
	"signalbot_go/modules/tv/internal/show"
	"signalbot_go/modules/tv/scrapers"
	"testing"
	"time"
)

func TestRtl(t *testing.T) {
	log := nopLog()

	scraper := &scrapers.Rtl{ScraperBase: scrapers.NewScraperBase(log, "rtl", location)}

	channel := make(chan show.Show)
	now := time.Date(2023, 4, 19, 0, 0, 0, 0, location)

	resp, err := os.Open("rtl_test.html")
	if err != nil {
		panic(err)
	}
	defer resp.Close()

	go scraper.Parse(resp, channel, now)

	// collect shows in list
	ss := []show.Show{}
	for s := range channel {
		ss = append(ss, s)
	}

	sendings := []show.Show{
		// 0
		{
			Date: time.Date(2023, 4, 19, 0, 0, 0, 0, location),
			Name: "RTL Nachtjournal",
		},
		// 1
		{
			Date: time.Date(2023, 4, 19, 0, 33, 0, 0, location),
			Name: "RTL Nachtjournal - Das Wetter",
		},
		// 2
		{
			Date: time.Date(2023, 4, 19, 0, 35, 0, 0, location),
			Name: "Colossos - Der Achterbahn-Gigant",
		},
		// 3
		{
			Date: time.Date(2023, 4, 19, 1, 25, 0, 0, location),
			Name: "CSI: Miami",
		},
		// 4
		{
			Date: time.Date(2023, 4, 19, 2, 15, 0, 0, location),
			Name: "CSI: Miami",
		},
		// 5
		{
			Date: time.Date(2023, 4, 19, 3, 0, 0, 0, location),
			Name: "CSI: Miami",
		},
		// 6
		{
			Date: time.Date(2023, 4, 19, 3, 55, 0, 0, location),
			Name: "CSI: Vegas",
		},
		// 7
		{
			Date: time.Date(2023, 4, 19, 4, 40, 0, 0, location),
			Name: "CSI: Vegas",
		},
		// 8
		{
			Date: time.Date(2023, 4, 19, 5, 20, 0, 0, location),
			Name: "CSI: Vegas",
		},
		// 9
		{
			Date: time.Date(2023, 4, 19, 6, 0, 0, 0, location),
			Name: "Punkt 6",
		},
		// 10
		{
			Date: time.Date(2023, 4, 19, 7, 0, 0, 0, location),
			Name: "Punkt 7",
		},
		// 11
		{
			Date: time.Date(2023, 4, 19, 8, 0, 0, 0, location),
			Name: "Punkt 8",
		},
		// 12
		{
			Date: time.Date(2023, 4, 19, 9, 0, 0, 0, location),
			Name: "Gute Zeiten, schlechte Zeiten",
		},
		// 13
		{
			Date: time.Date(2023, 4, 19, 9, 30, 0, 0, location),
			Name: "Unter uns",
		},
		// 14
		{
			Date: time.Date(2023, 4, 19, 10, 0, 0, 0, location),
			Name: "Ulrich Wetzel - Das Strafgericht",
		},
		// 15
		{
			Date: time.Date(2023, 4, 19, 11, 0, 0, 0, location),
			Name: "Barbara Salesch - Das Strafgericht",
		},
		// 16
		{
			Date: time.Date(2023, 4, 19, 12, 0, 0, 0, location),
			Name: "Punkt 12 - Das RTL-Mittagsjournal",
		},
		// 17
		{
			Date: time.Date(2023, 4, 19, 15, 0, 0, 0, location),
			Name: "Barbara Salesch - Das Strafgericht",
		},
		// 18
		{
			Date: time.Date(2023, 4, 19, 16, 0, 0, 0, location),
			Name: "Ulrich Wetzel - Das Strafgericht",
		},
		// 19
		{
			Date: time.Date(2023, 4, 19, 17, 0, 0, 0, location),
			Name: "RTL Aktuell",
		},
		// 20
		{
			Date: time.Date(2023, 4, 19, 17, 7, 0, 0, location),
			Name: "Explosiv Stories",
		},
		// 21
		{
			Date: time.Date(2023, 4, 19, 17, 30, 0, 0, location),
			Name: "Unter uns",
		},
		// 22
		{
			Date: time.Date(2023, 4, 19, 18, 0, 0, 0, location),
			Name: "Explosiv - Das Magazin",
		},
		// 23
		{
			Date: time.Date(2023, 4, 19, 18, 30, 0, 0, location),
			Name: "Exclusiv - Das Starmagazin",
		},
		// 24
		{
			Date: time.Date(2023, 4, 19, 18, 45, 0, 0, location),
			Name: "RTL Aktuell",
		},
		// 25
		{
			Date: time.Date(2023, 4, 19, 19, 3, 0, 0, location),
			Name: "RTL Aktuell - Das Wetter",
		},
		// 26
		{
			Date: time.Date(2023, 4, 19, 19, 5, 0, 0, location),
			Name: "Alles was zählt",
		},
		// 27
		{
			Date: time.Date(2023, 4, 19, 19, 40, 0, 0, location),
			Name: "Gute Zeiten, schlechte Zeiten",
		},
		// 28
		{
			Date: time.Date(2023, 4, 19, 20, 15, 0, 0, location),
			Name: "Der Bachelor",
		},
		// 29
		{
			Date: time.Date(2023, 4, 19, 22, 15, 0, 0, location),
			Name: "RTL Direkt",
		},
		// 30
		{
			Date: time.Date(2023, 4, 19, 22, 35, 0, 0, location),
			Name: "stern TV",
		},
	}

	if len(ss) != len(sendings) {
		t.Fatalf("Wrong amount of shows read. %d (should: %d)", len(ss), len(sendings))
	}

	s := ss[0]
	s_ref := sendings[0]

	if s.Name != s_ref.Name {
		t.Fatalf("Wrong name. %s (should: %s)", s.Name, s_ref.Name)
	}
	if !s.Date.Equal(s_ref.Date) {
		t.Fatalf("Wrong date. %v (should: %v)", s.Date, s_ref.Date)
	}

	s = ss[17]
	s_ref = sendings[17]

	if s.Name != s_ref.Name {
		t.Fatalf("Wrong name. %s (should: %s)", s.Name, s_ref.Name)
	}
	if !s.Date.Equal(s_ref.Date) {
		t.Fatalf("Wrong date. %v (should: %v)", s.Date, s_ref.Date)
	}

	s = ss[21]
	s_ref = sendings[21]

	if s.Name != s_ref.Name {
		t.Fatalf("Wrong name. %s (should: %s)", s.Name, s_ref.Name)
	}
	if !s.Date.Equal(s_ref.Date) {
		t.Fatalf("Wrong date. %v (should: %v)", s.Date, s_ref.Date)
	}
}
