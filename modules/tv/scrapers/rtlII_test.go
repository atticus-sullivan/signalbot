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

func TestRTLII(t *testing.T) {
	log := nopLog()

	scraper := &scrapers.Rtl2{ScraperBase: scrapers.NewScraperBase(log, "rtl2", location)}

	channel := make(chan show.Show)
	now := time.Date(2023, 4, 20, 0, 0, 0, 0, location)

	resp, err := os.Open("rtlII_test.html")
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
			Date: time.Date(2023, 4, 20, 5, 5, 0, 0, location),
			Name: "Der Trödeltrupp - Das Geld liegt im Keller · Sükrü bei Karin",
		},
		// 1
		{
			Date: time.Date(2023, 4, 20, 5, 55, 0, 0, location),
			Name: "Der Trödeltrupp - Das Geld liegt im Keller · Otto bei Asmus",
		},
		// 2
		{
			Date: time.Date(2023, 4, 20, 6, 55, 0, 0, location),
			Name: "Der Trödeltrupp - Das Geld liegt im Keller · Mauro bei Bruno, Lilly, Ute und Christian",
		},
		// 3
		{
			Date: time.Date(2023, 4, 20, 7, 55, 0, 0, location),
			Name: "Der Trödeltrupp - Das Geld liegt im Keller · Mauro, Otto und Sükrü bei Anna",
		},
		// 4
		{
			Date: time.Date(2023, 4, 20, 9, 55, 0, 0, location),
			Name: "Frauentausch · Sandy tauscht mit Marga",
		},
		// 5
		{
			Date: time.Date(2023, 4, 20, 11, 55, 0, 0, location),
			Name: "Frauentausch · Feten tauscht mit Heike",
		},
		// 6
		{
			Date: time.Date(2023, 4, 20, 13, 55, 0, 0, location),
			Name: "Hartz und herzlich · Die Plattenbauten von Bitterfeld-Wolfen (1)",
		},
		// 7
		{
			Date: time.Date(2023, 4, 20, 16, 0, 0, 0, location),
			Name: "RTLZWEI News · Episode 1068",
		},
		// 8
		{
			Date: time.Date(2023, 4, 20, 16, 4, 0, 0, location),
			Name: "RTLZWEI Wetter · Episode 1515",
		},
		// 9
		{
			Date: time.Date(2023, 4, 20, 16, 5, 0, 0, location),
			Name: "Hartz und herzlich - Tag für Tag Benz-Baracken · Zwischen Freud und Leid",
		},
		// 10
		{
			Date: time.Date(2023, 4, 20, 17, 5, 0, 0, location),
			Name: "Südklinik am Ring · Die Löwenmutter",
		},
		// 11
		{
			Date: time.Date(2023, 4, 20, 18, 5, 0, 0, location),
			Name: "Köln 50667 · Zusammen ist man weniger allein",
		},
		// 12
		{
			Date: time.Date(2023, 4, 20, 19, 5, 0, 0, location),
			Name: "Berlin - Tag & Nacht · Zurück in der Welt der Lebenden",
		},
		// 13
		{
			Date: time.Date(2023, 4, 20, 20, 15, 0, 0, location),
			Name: "Nachtschicht: Einsatz für die Lebensretter · Episode 008",
		},
		// 14
		{
			Date: time.Date(2023, 4, 20, 22, 15, 0, 0, location),
			Name: "Polizei im Einsatz · Eingesperrt im Park",
		},
		// 15
		{
			Date: time.Date(2023, 4, 20, 0, 20, 0, 0, location),
			Name: "Der Trödeltrupp - Das Geld liegt im Keller · Otto bei Christel",
		},
		// 16
		{
			Date: time.Date(2023, 4, 20, 1, 10, 0, 0, location),
			Name: "Der Trödeltrupp - Das Geld liegt im Keller · Mauro bei Bernie",
		},
		// 17
		{
			Date: time.Date(2023, 4, 20, 1, 55, 0, 0, location),
			Name: "Der Trödeltrupp - Das Geld liegt im Keller · Otto bei Susanne",
		},
		// 18
		{
			Date: time.Date(2023, 4, 20, 2, 40, 0, 0, location),
			Name: "Der Trödeltrupp - Das Geld liegt im Keller · Mauro bei Albrecht",
		},
		// 19
		{
			Date: time.Date(2023, 4, 20, 3, 25, 0, 0, location),
			Name: "Der Trödeltrupp - Das Geld liegt im Keller · Mauro bei Stefan",
		},
		// 20
		{
			Date: time.Date(2023, 4, 20, 4, 10, 0, 0, location),
			Name: "Der Trödeltrupp - Das Geld liegt im Keller · Mauro bei Manfred B.",
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

	s = ss[8]
	s_ref = sendings[8]

	if s.Name != s_ref.Name {
		t.Fatalf("Wrong name. %s (should: %s)", s.Name, s_ref.Name)
	}
	if !s.Date.Equal(s_ref.Date) {
		t.Fatalf("Wrong date. %v (should: %v)", s.Date, s_ref.Date)
	}

	s = ss[14]
	s_ref = sendings[14]

	if s.Name != s_ref.Name {
		t.Fatalf("Wrong name. %s (should: %s)", s.Name, s_ref.Name)
	}
	if !s.Date.Equal(s_ref.Date) {
		t.Fatalf("Wrong date. %v (should: %v)", s.Date, s_ref.Date)
	}
}
