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

func TestArd2_br(t *testing.T) {
	log := nopLog()

	scraper := &scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "arte", location), Url: "https://programm-api.ard.de/program/api/program?mode=channel&channelIds=arte"}

	channel := make(chan show.Show)
	now := time.Date(2023, 4, 19, 0, 0, 0, 0, location)

	resp, err := os.Open("ard2_test.json")
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
		{
			Date: time.Date(2024, 10, 19, 5, 30, 0, 0, location),
			Name: "Unesco-Weltkulturerbe - Schätze für die Ewigkeit -- Granada",
		},
		{
			Date: time.Date(2024, 10, 19, 6, 20, 0, 0, location),
			Name: "Gaudi - Architekt der Moderne in Barcelona -- Frankreich 2022",
		},
		{
			Date: time.Date(2024, 10, 19, 7, 15, 0, 0, location),
			Name: "360° Reportage -- La Réunion: Die Wiederbelebung der kreolischen Gärten",
		},
		{
			Date: time.Date(2024, 10, 19, 7, 50, 0, 0, location),
			Name: "Geo Reportage -- Yoga, Indiens erstaunliche Medizin",
		},
		{
			Date: time.Date(2024, 10, 19, 8, 45, 0, 0, location),
			Name: "GEO Reportage -- Die blinde Primaballerina von Sao Paulo",
		},
		{
			Date: time.Date(2024, 10, 19, 9, 40, 0, 0, location),
			Name: "Stadt Land Kunst Spezial -- Kenia",
		},
		{
			Date: time.Date(2024, 10, 19, 10, 20, 0, 0, location),
			Name: "Stadt Land Kunst Spezial -- Martinique",
		},
		{
			Date: time.Date(2024, 10, 19, 11, 0, 0, 0, location),
			Name: "Zu Tisch -- Garfagnana, Italien",
		},
		{
			Date: time.Date(2024, 10, 19, 11, 25, 0, 0, location),
			Name: "Im Bauch von Ljubljana -- Der Zentralmarkt",
		},
		{
			Date: time.Date(2024, 10, 19, 12, 20, 0, 0, location),
			Name: "Wasserlöcher - Oasen für Afrikas Fauna (1/3) -- Großbritannien 2020",
		},
		{
			Date: time.Date(2024, 10, 19, 13, 5, 0, 0, location),
			Name: "Wasserlöcher - Oasen für Afrikas Fauna (2/3) -- Großbritannien 2020",
		},
		{
			Date: time.Date(2024, 10, 19, 13, 50, 0, 0, location),
			Name: "Wasserlöcher - Oasen für Afrikas Fauna (3/3) -- Großbritannien 2020",
		},
		{
			Date: time.Date(2024, 10, 19, 14, 35, 0, 0, location),
			Name: "Pompeji, Geschichte einer Katastrophe (1/3) -- Im Schatten des Vesuv",
		},
		{
			Date: time.Date(2024, 10, 19, 15, 30, 0, 0, location),
			Name: "Pompeji, Geschichte einer Katastrophe (2/3) -- Flucht und Neuanfang",
		},
		{
			Date: time.Date(2024, 10, 19, 16, 30, 0, 0, location),
			Name: "Pompeji, Geschichte einer Katastrophe (3/3) -- Die letzten Stunden",
		},
		{
			Date: time.Date(2024, 10, 19, 17, 25, 0, 0, location),
			Name: "ARTE Reportage -- Libanon / Israel",
		},
		{
			Date: time.Date(2024, 10, 19, 18, 20, 0, 0, location),
			Name: "Mit offenen Karten -- Elektroautops - Wer stoppt China?",
		},
		{
			Date: time.Date(2024, 10, 19, 18, 35, 0, 0, location),
			Name: "Die letzten Venezianer -- Italien 2021",
		},
		{
			Date: time.Date(2024, 10, 19, 19, 20, 0, 0, location),
			Name: "ARTE Journal -- Die Abendausgabe des europäischen Nachrichtenmagazins",
		},
		{
			Date: time.Date(2024, 10, 19, 19, 40, 0, 0, location),
			Name: "360° Reportage -- Mongolei: Der Pferderetter",
		},
		{
			Date: time.Date(2024, 10, 19, 20, 15, 0, 0, location),
			Name: "Sardinien - Das Rätsel der Nuraghen-Türme -- Frankreich 2024",
		},
		{
			Date: time.Date(2024, 10, 19, 21, 45, 0, 0, location),
			Name: "Superfood Bohnen -- Deutschland 2024",
		},
		{
			Date: time.Date(2024, 10, 19, 22, 40, 0, 0, location),
			Name: "Unser Bauch - Die wunderbare Welt des Mikrobioms -- Frankreich 2019",
		},
		{
			Date: time.Date(2024, 10, 19, 23, 40, 0, 0, location),
			Name: "Muss Wohnen so teuer sein? -- 42 - Die Antwort auf fast alles",
		},
		{
			Date: time.Date(2024, 10, 20, 0, 10, 0, 0, location),
			Name: "Kurzschluss -- Schwer verliebt",
		},
		{
			Date: time.Date(2024, 10, 20, 1, 5, 0, 0, location),
			Name: "Antoine und Colette -- Spielfilm Frankreich 1961",
		},
		{
			Date: time.Date(2024, 10, 20, 1, 40, 0, 0, location),
			Name: "Sonnenstürme - Die rätselhafte Gefahr -- Deutschland 2020",
		},
		{
			Date: time.Date(2024, 10, 20, 2, 30, 0, 0, location),
			Name: "Verschollene Filmschätze -- 1938. Chamberlains Treffen mit Hitler",
		},
		{
			Date: time.Date(2024, 10, 20, 3, 0, 0, 0, location),
			Name: "28 Minuten -- Frankreich 2024",
		},
	}

	if len(ss) != len(sendings) { // 4 after 00:00
		t.Fatalf("Wrong amount of shows read. %d (should: %d)", len(ss), len(sendings))
	}

	for i, s := range ss {
		s_ref := sendings[i]
		if s.Name != s_ref.Name {
			t.Fatalf("Wrong name. %s (should: %s)", s.Name, s_ref.Name)
		}
		if !s.Date.Equal(s_ref.Date) {
			t.Fatalf("Wrong date. %v (should: %v)", s.Date, s_ref.Date)
		}
	}
}
