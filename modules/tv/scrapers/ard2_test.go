package scrapers_test

import (
	"fmt"
	"os"
	"signalbot_go/modules/tv/internal/show"
	"signalbot_go/modules/tv/scrapers"
	"testing"
	"time"

	"golang.org/x/exp/slog"
)

func TestTestArd2(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout))
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}

	scraper := &scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "br", loc), Url: "https://programm.ard.de/TV/Programm/Sender?sender=-28107&datum=%s&hour=0&archiv=1"}

	channel := make(chan show.Show)
	now := time.Now()

	resp, err := scraper.Get(now)
	if err != nil {
		panic(err)
	}
	defer resp.Close()

	go scraper.Parse(resp, channel, now)

	for s := range channel {
		fmt.Println(s)
	}
}
