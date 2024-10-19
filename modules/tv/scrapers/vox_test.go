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

func TestVox(t *testing.T) {
	log := nopLog()

	scraper := &scrapers.Vox{ScraperBase: scrapers.NewScraperBase(log, "vox", location)}

	channel := make(chan show.Show)
	now := time.Date(2023, 4, 20, 0, 0, 0, 0, location)

	resp, err := os.Open("vox_test.html")
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
			Date: time.Date(2023, 4, 20, 0, 0, 0, 0, location),
			Name: "vox nachrichten",
		},
		// 1
		{
			Date: time.Date(2023, 4, 20, 0, 20, 0, 0, location),
			Name: "Medical Detectives - Geheimnisse der Gerichtsmedizin",
		},
		// 2
		{
			Date: time.Date(2023, 4, 20, 1, 20, 0, 0, location),
			Name: "Medical Detectives - Geheimnisse der Gerichtsmedizin",
		},
		// 3
		{
			Date: time.Date(2023, 4, 20, 2, 10, 0, 0, location),
			Name: "Medical Detectives - Geheimnisse der Gerichtsmedizin",
		},
		// 4
		{
			Date: time.Date(2023, 4, 20, 3, 0, 0, 0, location),
			Name: "Medical Detectives - Geheimnisse der Gerichtsmedizin",
		},
		// 5
		{
			Date: time.Date(2023, 4, 20, 3, 50, 0, 0, location),
			Name: "Medical Detectives - Geheimnisse der Gerichtsmedizin",
		},
		// 6
		{
			Date: time.Date(2023, 4, 20, 4, 40, 0, 0, location),
			Name: "Medical Detectives - Geheimnisse der Gerichtsmedizin",
		},
		// 7
		{
			Date: time.Date(2023, 4, 20, 5, 0, 0, 0, location),
			Name: "CSI: NY",
		},
		// 8
		{
			Date: time.Date(2023, 4, 20, 5, 45, 0, 0, location),
			Name: "CSI: NY",
		},
		// 9
		{
			Date: time.Date(2023, 4, 20, 6, 30, 0, 0, location),
			Name: "CSI: NY",
		},
		// 10
		{
			Date: time.Date(2023, 4, 20, 7, 20, 0, 0, location),
			Name: "CSI: Den Tätern auf der Spur",
		},
		// 11
		{
			Date: time.Date(2023, 4, 20, 8, 10, 0, 0, location),
			Name: "CSI: Den Tätern auf der Spur",
		},
		// 12
		{
			Date: time.Date(2023, 4, 20, 9, 10, 0, 0, location),
			Name: "CSI: Miami",
		},
		// 13
		{
			Date: time.Date(2023, 4, 20, 10, 0, 0, 0, location),
			Name: "CSI: Miami",
		},
		// 14
		{
			Date: time.Date(2023, 4, 20, 11, 0, 0, 0, location),
			Name: "CSI: Miami",
		},
		// 15
		{
			Date: time.Date(2023, 4, 20, 11, 55, 0, 0, location),
			Name: "vox nachrichten",
		},
		// 16
		{
			Date: time.Date(2023, 4, 20, 12, 0, 0, 0, location),
			Name: "Shopping Queen",
		},
		// 17
		{
			Date: time.Date(2023, 4, 20, 13, 0, 0, 0, location),
			Name: "Zwischen Tüll und Tränen",
		},
		// 18
		{
			Date: time.Date(2023, 4, 20, 14, 0, 0, 0, location),
			Name: "Full House - Familie XXL",
		},
		// 19
		{
			Date: time.Date(2023, 4, 20, 15, 0, 0, 0, location),
			Name: "Shopping Queen",
		},
		// 20
		{
			Date: time.Date(2023, 4, 20, 16, 0, 0, 0, location),
			Name: "Das Duell - Zwischen Tüll und Tränen",
		},
		// 21
		{
			Date: time.Date(2023, 4, 20, 17, 0, 0, 0, location),
			Name: "Zwischen Tüll und Tränen",
		},
		// 22
		{
			Date: time.Date(2023, 4, 20, 18, 0, 0, 0, location),
			Name: "First Dates - Ein Tisch für zwei",
		},
		// 23
		{
			Date: time.Date(2023, 4, 20, 19, 0, 0, 0, location),
			Name: "Das perfekte Dinner",
		},
		// 24
		{
			Date: time.Date(2023, 4, 20, 20, 15, 0, 0, location),
			Name: "Sing",
		},
		// 25
		{
			Date: time.Date(2023, 4, 20, 22, 25, 0, 0, location),
			Name: "Bad Boys 2",
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
