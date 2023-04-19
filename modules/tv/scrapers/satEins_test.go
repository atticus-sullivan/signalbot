package scrapers_test

import (
	"os"
	"signalbot_go/modules/tv/internal/show"
	"signalbot_go/modules/tv/scrapers"
	"testing"
	"time"
)

func TestSatEins_sat1(t *testing.T) {
	log := nopLog()

	scraper := &scrapers.SatEins{ScraperBase: scrapers.NewScraperBase(log, "sat1", location), Url: "https://www.sat1.de/tv-programm"}

	channel := make(chan show.Show)
	now := time.Date(2023, 4, 19, 0, 0, 0, 0, location)

	resp, err := os.Open("satEins-sat1_test.html")
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
			Date: time.Date(2023, 4, 19, 5, 30, 0, 0, location),
			Name: "SAT.1-Frühstücksfernsehen · SAT.1-FRÜHSTÜCKSFERNSEHEN 2023 - FOLGE 076 · S2023, E76",
		},
		// 1
		{
			Date: time.Date(2023, 4, 19, 10, 0, 0, 0, location),
			Name: "Die Ruhrpottwache - Vermisstenfahnder im Einsatz · Zugfahrt ins Nirgendwo · S02, E63",
		},
		// 2
		{
			Date: time.Date(2023, 4, 19, 10, 30, 0, 0, location),
			Name: "Die Ruhrpottwache - Vermisstenfahnder im Einsatz · Vergiftete Liebe · S02, E55",
		},
		// 3
		{
			Date: time.Date(2023, 4, 19, 11, 0, 0, 0, location),
			Name: "Auf Streife - Die Spezialisten · Über den Dächern von Kölle · S07, E35",
		},
		// 4
		{
			Date: time.Date(2023, 4, 19, 12, 0, 0, 0, location),
			Name: "Auf Streife · Mutterschmutz · S09, E49",
		},
		// 5
		{
			Date: time.Date(2023, 4, 19, 13, 0, 0, 0, location),
			Name: "Auf Streife · Du bist nicht meine Mutter · S10, E20",
		},
		// 6
		{
			Date: time.Date(2023, 4, 19, 14, 0, 0, 0, location),
			Name: "Klinik am Südring · Mein Bruderherz · S03, E45",
		},
		// 7
		{
			Date: time.Date(2023, 4, 19, 15, 0, 0, 0, location),
			Name: "Klinik am Südring · Ich war noch niemals in Paris · S03, E69",
		},
		// 8
		{
			Date: time.Date(2023, 4, 19, 16, 0, 0, 0, location),
			Name: "Volles Haus! SAT.1 Live / SAT.1 Regional-Magazine um 17:30 Uhr · S01",
		},
		// 9
		{
			Date: time.Date(2023, 4, 19, 19, 0, 0, 0, location),
			Name: "Die perfekte Minute · S01",
		},
		// 10
		{
			Date: time.Date(2023, 4, 19, 19, 55, 0, 0, location),
			Name: "SAT.1 Nachrichten · S2023",
		},
		// 11
		{
			Date: time.Date(2023, 4, 19, 20, 15, 0, 0, location),
			Name: "Das 1% Quiz - Wie clever ist Deutschland? · S01",
		},
		// 12
		{
			Date: time.Date(2023, 4, 19, 22, 15, 0, 0, location),
			Name: "Die perfekte Minute · Die perfekte Minute - LARISSA UND RALF · S01, E11",
		},
		// 13
		{
			Date: time.Date(2023, 4, 19, 23, 15, 0, 0, location),
			Name: "Die perfekte Minute · Die perfekte Minute - ROBIN UND LUUK · S01, E12",
		},
		// 14
		{
			Date: time.Date(2023, 4, 19, 0, 10, 0, 0, location),
			Name: "Das 1% Quiz - Wie clever ist Deutschland? · S01",
		},
		// 15
		{
			Date: time.Date(2023, 4, 19, 1, 50, 0, 0, location),
			Name: "Auf Streife - Die Spezialisten · Sandras Rache · E46",
		},
		// 16
		{
			Date: time.Date(2023, 4, 19, 2, 35, 0, 0, location),
			Name: "Auf Streife - Die Spezialisten · Vergaloppiert · E42",
		},
		// 17
		{
			Date: time.Date(2023, 4, 19, 3, 20, 0, 0, location),
			Name: "Auf Streife - Die Spezialisten · Lenis Geheimnis · E10",
		},
		// 18
		{
			Date: time.Date(2023, 4, 19, 4, 0, 0, 0, location),
			Name: "Auf Streife - Die Spezialisten · Alte Liebe, neu entflammt · E13",
		},
		// 19
		{
			Date: time.Date(2023, 4, 19, 4, 45, 0, 0, location),
			Name: "Auf Streife · Spa-Maßnahmen · S09, E48",
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
