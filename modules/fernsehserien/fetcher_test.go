package fernsehserien

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
	"io"
	"os"
	"testing"
	"time"

	"log/slog"
)

func nopLog() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
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
	fC, err := os.Open("test1C.html")
	if err != nil {
		panic(err)
	}
	defer fC.Close()
	items, err := fserie.fetcher.getFromReaders(map[string]io.ReadCloser{"nameNothing": fA, "name": fB, "tatort3": fC}, fserie.UnavailableSenders)
	if err != nil {
		panic(err)
	}

	if len(items) != 4 {
		t.Fatalf("Invalid amount of items found was %d (should %d)", len(items), 4)
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

	i = items[2]
	i_ref = sending{
		Date:   time.Date(2024, 9, 6, 22, 50, 0, 0, location),
		Sender: "Das Erste",
		Name:   "tatort3",
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

	i = items[3]
	i_ref = sending{
		Date:   time.Date(2024, 9, 7, 2, 0, 0, 0, location),
		Sender: "Das Erste",
		Name:   "tatort3",
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
