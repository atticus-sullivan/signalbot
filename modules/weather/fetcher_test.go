package weather

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
