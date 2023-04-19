package scrapers_test

import (
	"os"
	"signalbot_go/modules/tv/internal/show"
	"signalbot_go/modules/tv/scrapers"
	"testing"
	"time"
)

func TestArd2_br(t *testing.T) {
	log := nopLog()

	scraper := &scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "br", location), Url: "https://programm.ard.de/TV/Programm/Sender?sender=-28107&datum=%s&hour=0&archiv=1"}

	channel := make(chan show.Show)
	now := time.Date(2023, 4, 19, 0, 0, 0, 0, location)

	resp, err := os.Open("ard2-br_test.html")
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
			Date: time.Date(2023, 4, 19, 6, 0, 0, 0, location),
			Name: "Dahoam is Dahoam (3140) · Erhitzte Gemüter",
		},
		// 1
		{
			Date: time.Date(2023, 4, 19, 6, 30, 0, 0, location),
			Name: "Sturm der Liebe (4014) Der Antrag",
		},
		// 2
		{
			Date: time.Date(2023, 4, 19, 7, 20, 0, 0, location),
			Name: "Tele-Gym Integrales Qi Gong",
		},
		// 3
		{
			Date: time.Date(2023, 4, 19, 7, 35, 0, 0, location),
			Name: "Panoramabilder/Bergwetter mit Nachrichten aus dem Bayerntext",
		},
		// 4
		{
			Date: time.Date(2023, 4, 19, 8, 40, 0, 0, location),
			Name: "Tele-Gym Schlank, Fit & Gesund",
		},
		// 5
		{
			Date: time.Date(2023, 4, 19, 8, 55, 0, 0, location),
			Name: "Panoramabilder/Bergwetter mit Nachrichten aus dem Bayerntext",
		},
		// 6
		{
			Date: time.Date(2023, 4, 19, 9, 10, 0, 0, location),
			Name: "Eisbär, Affe & Co Zoogeschichten aus Stuttgart",
		},
		// 7
		{
			Date: time.Date(2023, 4, 19, 10, 0, 0, 0, location),
			Name: "Giraffe, Erdmännchen & Co Zoogeschichten aus Frankfurt und Kronberg",
		},
		// 8
		{
			Date: time.Date(2023, 4, 19, 10, 50, 0, 0, location),
			Name: "Welt der Tiere Igel unter uns",
		},
		// 9
		{
			Date: time.Date(2023, 4, 19, 11, 20, 0, 0, location),
			Name: "Abenteuer Wildnis Die Leopardin - Gejagte Jägerin",
		},
		// 10
		{
			Date: time.Date(2023, 4, 19, 12, 5, 0, 0, location),
			Name: "nah und fern Tabernas | Großer Falkenstein | Zukunftsmuseum Nürnberg",
		},
		// 11
		{
			Date: time.Date(2023, 4, 19, 12, 35, 0, 0, location),
			Name: "Gefragt - Gejagt Moderation: Alexander Bommes",
		},
		// 12
		{
			Date: time.Date(2023, 4, 19, 13, 20, 0, 0, location),
			Name: "Quizduell-Olymp Moderation: Esther Sedlaczek",
		},
		// 13
		{
			Date: time.Date(2023, 4, 19, 14, 10, 0, 0, location),
			Name: "aktiv und gesund",
		},
		// 14
		{
			Date: time.Date(2023, 4, 19, 14, 40, 0, 0, location),
			Name: "Nashorn, Zebra & Co Zoogeschichten aus München",
		},
		// 15
		{
			Date: time.Date(2023, 4, 19, 15, 30, 0, 0, location),
			Name: "Schnittgut. Alles aus dem Garten",
		},
		// 16
		{
			Date: time.Date(2023, 4, 19, 16, 0, 0, 0, location),
			Name: "BR24 Nachrichten - Berichte - Wettervorhersage",
		},
		// 17
		{
			Date: time.Date(2023, 4, 19, 16, 15, 0, 0, location),
			Name: "Wir in Bayern · Lust auf Heimat",
		},
		// 18
		{
			Date: time.Date(2023, 4, 19, 17, 30, 0, 0, location),
			Name: "Regionalprogramm Frankenschau aktuell (BR Nord) | Abendschau - Der Süden (BFS Süd)",
		},
		// 19
		{
			Date: time.Date(2023, 4, 19, 18, 0, 0, 0, location),
			Name: "Abendschau · Das bewegt Bayern heute",
		},
		// 20
		{
			Date: time.Date(2023, 4, 19, 18, 30, 0, 0, location),
			Name: "BR24 Nachrichten - Berichte - Wettervorhersage",
		},
		// 21
		{
			Date: time.Date(2023, 4, 19, 19, 0, 0, 0, location),
			Name: "STATIONEN Teufel, komm raus!",
		},
		// 22
		{
			Date: time.Date(2023, 4, 19, 19, 30, 0, 0, location),
			Name: "Dahoam is Dahoam (3141) Für meine Tochter nur das Beste",
		},
		// 23
		{
			Date: time.Date(2023, 4, 19, 20, 0, 0, 0, location),
			Name: "Tagesschau",
		},
		// 24
		{
			Date: time.Date(2023, 4, 19, 20, 15, 0, 0, location),
			Name: "Münchner Runde Freizeit statt Karriere – Lohnt sich Leistung noch?",
		},
		// 25
		{
			Date: time.Date(2023, 4, 19, 21, 15, 0, 0, location),
			Name: "Kontrovers Moderation: Ursula Heller",
		},
		// 26
		{
			Date: time.Date(2023, 4, 19, 21, 45, 0, 0, location),
			Name: "BR24 Nachrichten - Berichte - Wettervorhersage",
		},
		// 27
		{
			Date: time.Date(2023, 4, 19, 22, 0, 0, 0, location),
			Name: "DokThema Deutschland schaltet ab - Der Atomausstieg und die Folgen",
		},
		// 27
		{
			Date: time.Date(2023, 4, 19, 22, 45, 0, 0, location),
			Name: "Wim Wenders, Desperado",
		},
		// 28
		{
			Date: time.Date(2023, 4, 19, 0, 45, 0, 0, location),
			Name: "kinokino",
		},
		// 29
		{
			Date: time.Date(2023, 4, 19, 1, 0, 0, 0, location),
			Name: "Am Ende der Gewalt Spielfilm Frankreich/Deutschland/USA 1997",
		},
		// 30
		{
			Date: time.Date(2023, 4, 19, 2, 55, 0, 0, location),
			Name: "Land of Plenty Spielfilm Deutschland/USA 2003",
		},
		// 31
		{
			Date: time.Date(2023, 4, 19, 4, 55, 0, 0, location),
			Name: "Abendschau Das bewegt Bayern heute",
		},
	}

	if len(ss) != len(sendings) { // 4 after 00:00
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

	s = ss[19]
	s_ref = sendings[19]

	if s.Name != s_ref.Name {
		t.Fatalf("Wrong name. %s (should: %s)", s.Name, s_ref.Name)
	}
	if !s.Date.Equal(s_ref.Date) {
		t.Fatalf("Wrong date. %v (should: %v)", s.Date, s_ref.Date)
	}
}
