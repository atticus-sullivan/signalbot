package weather

import (
	"io"
	"os"
	"testing"

	"log/slog"
)

func nopLog() *slog.Logger {
	return slog.New(slog.HandlerOptions{Level: slog.LevelError}.NewTextHandler(io.Discard))
}

func TestFetcher(t *testing.T) {
	log := nopLog()
	weather, err := NewWeather(log, "./")
	if err != nil {
		panic(err)
	}

	f, err := os.Open("test1.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	resp, err := weather.Fetcher.getFromReader(f)
	if err != nil {
		panic(err)
	}

	// fo,_ := os.Create("test1.out")
	// fo.Write([]byte(resp.String()))
	// fo.Close()

	str := resp.String()
	out, err := os.ReadFile("test1.out")
	if err != nil {
		panic(err)
	}

	if str != string(out) {
		t.Fatalf("formatting is wrong")
	}
}
