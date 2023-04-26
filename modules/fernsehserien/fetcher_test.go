package fernsehserien

import (
	"io"
	"os"
	"testing"
	"time"

	"golang.org/x/exp/slog"
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

func TestFetcher(t *testing.T) {
	log := nopLog()
	fserie, err := NewFernsehserien(log, "./")
	if err != nil {
		panic(err)
	}

	fA, err := os.Open("test1A.html")
	if err != nil {
		panic(err)
	}
	defer fA.Close()
	fB, err := os.Open("test1B.html")
	if err != nil {
		panic(err)
	}
	defer fB.Close()
	items, err := fserie.fetcher.getFromReaders(map[string]io.ReadCloser{"nameNothing": fA, "name": fB}, fserie.UnavailableSenders)
	if err != nil {
		panic(err)
	}

	if len(items) != 2 {
		t.Fatalf("Invalid amount of items found")
	}

	i := items[0]
	i_ref := sending{
		Date:   time.Date(2023, 4, 24, 20, 15, 0, 0, location),
		Sender: "Kabel Eins",
		Name:   "name",
	}
	if !i_ref.Date.Equal(i.Date) {
		t.Fatalf("Invalid date. Is %v (should: %v)", i.Date, i_ref.Date)
	}
	if i.Sender != i_ref.Sender {
		t.Fatalf("Invalid sender. Is %v (should: %v)", i.Sender, i_ref.Sender)
	}
	if i.Name != i_ref.Name {
		t.Fatalf("Invalid name")
	}

	i = items[1]
	i_ref = sending{
		Date:   time.Date(2023, 4, 25, 0, 25, 0, 0, location),
		Sender: "Kabel Eins",
		Name:   "name",
	}
	if !i_ref.Date.Equal(i.Date) {
		t.Fatalf("Invalid date. Is %v (should: %v)", i.Date, i_ref.Date)
	}
	if i.Sender != i_ref.Sender {
		t.Fatalf("Invalid sender. Is %v (should: %v)", i.Sender, i_ref.Sender)
	}
	if i.Name != i_ref.Name {
		t.Fatalf("Invalid name")
	}
}

// func TestMain(t *testing.T) {
// 	fserie, err := NewFernsehserien(slog.Default(), "./")
// 	if err != nil {
// 		panic(err)
// 	}
// 	reader, err := fserie.fetcher.getReaders(map[string]string{"test": "https://www.fernsehserien.de/filme/die-drei-musketiere-d-artagnan"})
// 	if err != nil {
// 		panic(err)
// 	}
// 	out, err := fserie.fetcher.getFromReaders(reader, map[string]bool{})
// 	if err != nil {
// 		panic(err)
// 	}
// 	t.Logf("%d \"%v\"\n", len(out), out)
// 	t.Fail()
// }
