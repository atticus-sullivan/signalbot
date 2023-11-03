package scrapers_test

import (
	"io"
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
