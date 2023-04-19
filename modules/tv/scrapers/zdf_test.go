package scrapers_test

import (
	"os"
	"signalbot_go/modules/tv/internal/show"
	"signalbot_go/modules/tv/scrapers"
	"testing"
	"time"

	"github.com/andybalholm/cascadia"
)

func TestZdf(t *testing.T) {
	log := nopLog()

	scraper := scrapers.NewZdf(scrapers.NewScraperBase(log, "zdf", location), cascadia.MustCompile(".timeline-ZDF"))

	channel := make(chan show.Show)
	now := time.Date(2023, 4, 20, 0, 0, 0, 0, location)

	resp, err := os.Open("zdf-zdf_test.html")
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
			Date: time.Date(2023, 4, 20, 5, 30, 0, 0, location),
			Name: "ZDF-Morgenmagazin",
		},
		{
			Date: time.Date(2023, 4, 20, 9, 0, 0, 0, location),
			Name: "heute Xpress",
		},
		{
			Date: time.Date(2023, 4, 20, 9, 5, 0, 0, location),
			Name: "Volle Kanne - Service täglich",
		},
		{
			Date: time.Date(2023, 4, 20, 10, 30, 0, 0, location),
			Name: "Notruf Hafenkante",
		},
		{
			Date: time.Date(2023, 4, 20, 11, 15, 0, 0, location),
			Name: "SOKO Wismar",
		},
		{
			Date: time.Date(2023, 4, 20, 12, 0, 0, 0, location),
			Name: "heute",
		},
		{
			Date: time.Date(2023, 4, 20, 12, 10, 0, 0, location),
			Name: "drehscheibe",
		},
		{
			Date: time.Date(2023, 4, 20, 13, 0, 0, 0, location),
			Name: "ZDF-Mittagsmagazin",
		},
		{
			Date: time.Date(2023, 4, 20, 14, 0, 0, 0, location),
			Name: "heute - in Deutschland",
		},
		{
			Date: time.Date(2023, 4, 20, 14, 15, 0, 0, location),
			Name: "Die Küchenschlacht",
		},
		{
			Date: time.Date(2023, 4, 20, 15, 0, 0, 0, location),
			Name: "heute Xpress",
		},
		{
			Date: time.Date(2023, 4, 20, 15, 5, 0, 0, location),
			Name: "Bares für Rares",
		},
		{
			Date: time.Date(2023, 4, 20, 16, 0, 0, 0, location),
			Name: "heute - in Europa",
		},
		{
			Date: time.Date(2023, 4, 20, 16, 10, 0, 0, location),
			Name: "Die Rosenheim-Cops",
		},
		{
			Date: time.Date(2023, 4, 20, 17, 0, 0, 0, location),
			Name: "heute",
		},
		{
			Date: time.Date(2023, 4, 20, 17, 10, 0, 0, location),
			Name: "hallo deutschland",
		},
		{
			Date: time.Date(2023, 4, 20, 17, 50, 0, 0, location),
			Name: "Leute heute",
		},
		{
			Date: time.Date(2023, 4, 20, 18, 5, 0, 0, location),
			Name: "SOKO Stuttgart",
		},
		{
			Date: time.Date(2023, 4, 20, 19, 0, 0, 0, location),
			Name: "heute",
		},
		{
			Date: time.Date(2023, 4, 20, 19, 20, 0, 0, location),
			Name: "Wetter",
		},
		{
			Date: time.Date(2023, 4, 20, 19, 25, 0, 0, location),
			Name: "Notruf Hafenkante",
		},
		{
			Date: time.Date(2023, 4, 20, 20, 15, 0, 0, location),
			Name: "Lena Lorenz - Lügenbaby",
		},
		{
			Date: time.Date(2023, 4, 20, 21, 45, 0, 0, location),
			Name: "heute journal",
		},
		{
			Date: time.Date(2023, 4, 20, 22, 15, 0, 0, location),
			Name: "maybrit illner",
		},
		{
			Date: time.Date(2023, 4, 20, 23, 15, 0, 0, location),
			Name: "Markus Lanz",
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
