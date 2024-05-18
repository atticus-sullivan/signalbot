package buechertreff

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

	"log/slog"
)

func nopLog() *slog.Logger {
	return slog.New(slog.HandlerOptions{Level: slog.LevelError}.NewTextHandler(io.Discard))
}

func TestFetcher(t *testing.T) {
	log := nopLog()
	buecher, err := NewBuechertreff(log, "./")
	if err != nil {
		panic(err)
	}

	f, err := os.Open("test1.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	books, err := buecher.fetcher.getFromReader(f)
	if err != nil {
		panic(err)
	}

	if len(books) != 2 {
		t.Fatalf("Invalid number of books parsed")
	}

	b := books[0]
	b_ref := bookItem{
		pos:      "Band 1",
		name:     "Tag der Geister",
		origName: "Zaklinateli,",
		pub:      "2011",
	}
	if b.pos != b_ref.pos {
		t.Fatalf("Wrong pos (0)")
	}
	if b.name != b_ref.name {
		t.Fatalf("Wrong name (0)")
	}
	if b.origName != b_ref.origName {
		t.Fatalf("Wrong origName. Was %s (should: %s)", b.origName, b_ref.origName)
	}
	if b.pub != b_ref.pub {
		t.Fatalf("Wrong pub (0)")
	}

	b = books[1]
	b_ref = bookItem{
		pos:      "Band 2",
		name:     "Turm des Ordens",
		origName: "Lovushka dlya dukha,",
		pub:      "2014",
	}
	if b.pos != b_ref.pos {
		t.Fatalf("Wrong pos (1)")
	}
	if b.name != b_ref.name {
		t.Fatalf("Wrong name (1)")
	}
	if b.origName != b_ref.origName {
		t.Fatalf("Wrong origName. Was %s (should: %s)", b.origName, b_ref.origName)
	}
	if b.pub != b_ref.pub {
		t.Fatalf("Wrong pub (1)")
	}

	// fo,_ := os.Create("test1.out")
	// fo.Write([]byte(books.String()))
	// fo.Close()

	str := books.String()
	out, err := os.ReadFile("test1.out")
	if err != nil {
		panic(err)
	}

	if str != string(out) {
		t.Fatalf("formatting is wrong")
	}
}
