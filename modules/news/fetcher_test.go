package news

import (
	"io"
	"os"
	"testing"
	"time"

	"log/slog"
)

func nopLog() *slog.Logger {
	return slog.New(slog.HandlerOptions{Level: slog.LevelError}.NewTextHandler(io.Discard))
}

func loadZone() *time.Location {
	if loc, err := time.LoadLocation("Europe/Berlin"); err != nil {
		panic(err)
	} else {
		return loc
	}
}

var location *time.Location = loadZone()

func TestFetcherNews(t *testing.T) {
	log := nopLog()
	news, err := NewNews(log, "./")
	if err != nil {
		panic(err)
	}

	f, err := os.Open("testNews1.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	items, err := news.fetcher.getNewsFromReader(f)
	if err != nil {
		panic(err)
	}

	if len(items) != 10 {
		t.Fatalf("Invalid amount of news read")
	}

	n := items[0]
	n_ref := entry{
		Date:          time.Date(2023, 04, 19, 12, 56, 44, 907000000, location),
		Webpage:       "https://www.tagesschau.de/inland/gesellschaft/gebaeudeenergiegesetz-pressekonferenz-101.html",
		Title:         "Kabinett beschließt Pläne zum Heizungstausch",
		Topline:       "Novelle des Gebäudeenergiegesetzes",
		FirstSentence: "Bundeswirtschaftsminister Habeck und Bundesbauministerin Geywitz stellen ein begleitendes Förderkonzept vor.",
		Content: []contentLine{
			{
				Value: "<strong>Die Bundesregierung hat sich auf eine Novelle des Gebäudeenergiegesetzes geeinigt. In einer Pressekonferenz stellten Bundeswirtschaftsminister Habeck (Grüne) und Bundesbauministerin Geywitz (SPD) auch ein begleitendes Förderkonzept vor.</strong>",
				Type:  "text",
			},
			{
				Value: "Das Bundeskabinett hat in Berlin den Gesetzentwurf für die Umstellung von Heizungen auf erneuerbare Energien gebilligt. Danach sollen vom kommenden Jahr an alle neu eingebauten Heizungen zu mindestens 65 Prozent mit erneuerbaren Energien betrieben werden. ",
				Type:  "text",
			},
			{
				Value: "Die Vorschriften werden zur Vermeidung sozialer Härten von Ausnahmen, Übergangsregelungen und Förderungsmöglichkeiten flankiert. Bundeswirtschaftsminister Robert Habeck (Grüne) und Bundesbauministerin Klara Geywitz stellen die Gesetzesnovelle zurzeit auf einer Pressekonferenz vor.",
				Type:  "text",
			},
			{
				Value: "<em>Weitere Informationen in Kürze.</em>",
				Type:  "text",
			},
			{
				Value: "",
				Type:  "box",
			},
		},
	}
	if !n.Date.Equal(n_ref.Date) {
		t.Fatalf("Wrong date parsed %v (should: %v)", n.Date, n_ref.Date)
	}
	if n.Webpage != n_ref.Webpage {
		t.Fatalf("Wrong Webpage")
	}
	if n.Title != n_ref.Title {
		t.Fatalf("Wrong Title")
	}
	if n.Topline != n_ref.Topline {
		t.Fatalf("Wrong Topline")
	}
	if n.FirstSentence != n_ref.FirstSentence {
		t.Fatalf("Wrong FirstSentence %s (should: %s)", n.FirstSentence, n_ref.FirstSentence)
	}
	if len(n.Content) != len(n_ref.Content) {
		t.Fatalf("length of content is wrong")
	}
	for i := range n.Content {
		if n.Content[i] != n_ref.Content[i] {
			t.Fatalf("Content item is wrong %+v (should: %+v)", n.Content[i], n_ref.Content[i])
		}
	}

	// fo,_ := os.Create("testNews1.out")
	// fo.Write([]byte(news.String()))
	// fo.Close()

	str := items.String()
	out, err := os.ReadFile("testNews1.out")
	if err != nil {
		panic(err)
	}

	if str != string(out) {
		t.Fatalf("formatting is wrong")
	}
}

func TestFetcherBreakingNeg(t *testing.T) {
	log := nopLog()
	news, err := NewNews(log, "./")
	if err != nil {
		panic(err)
	}

	f, err := os.Open("testBreaking1.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	items, err := news.fetcher.getBreakingFromReader(f)
	if err != nil {
		panic(err)
	}

	if len(items) != 0 {
		t.Fatalf("Invalid amount of breakings, none should be present")
	}
}

func TestFetcherBreakingPos(t *testing.T) {
	log := nopLog()
	news, err := NewNews(log, "./")
	if err != nil {
		panic(err)
	}

	f, err := os.Open("testBreaking1B.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	items, err := news.fetcher.getBreakingFromReader(f)
	if err != nil {
		panic(err)
	}

	if len(items) != 1 {
		t.Fatalf("Invalid amount of breakings, none should be present")
	}
	n := items[0]
	n_ref := breaking{
		Headline: "Bürgermeisterwahl in Berlin: Wegner im 1. Wahlgang gescheitert",
		Text:     "Der CDU-Politiker Kai Wegner ist bei der Wahl zum neuen Regierenden Bürgermeister von Berlin im ersten Wahlgang gescheitert. Der 50-Jährige verpasste bei der Abstimmung im Abgeordnetenhaus die erforderliche absolute Mehrheit.",
		Url:      "/eilmeldung/eilmeldung-7196.html",
		LinkText: "Eilmeldung",
		Id:       "54ec4992-8048-4308-b5e4-4541e79dfc2f",
		Date:     time.Date(2023, 4, 27, 12, 53, 0, 0, location),
		DateS: "Stand: 27.04.2023 12:53 Uhr",
	}
	if n.Headline != n_ref.Headline {
		t.Fatalf("Wrong Headline")
	}
	if n.Text != n_ref.Text {
		t.Fatalf("Wrong Text")
	}
	if n.Url != n_ref.Url {
		t.Fatalf("Wrong Url")
	}
	if n.LinkText != n_ref.LinkText {
		t.Fatalf("Wrong LinkText")
	}
	if n.Id != n_ref.Id {
		t.Fatalf("Wrong Id")
	}
	if n.DateS != n_ref.DateS {
		t.Fatalf("Wrong dateS")
	}
	if !n.Date.Equal(n_ref.Date) {
		t.Fatalf("Wrong date. Is %v (should: %v)", n.Date, n_ref.Date)
	}
}
