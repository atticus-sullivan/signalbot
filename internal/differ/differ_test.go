package differ_test

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
	"signalbot_go/internal/differ"
	"testing"
)

type content string

func (c content) AddString() string {
	return "> " + string(c)
}
func (c content) RemString() string {
	return "< " + string(c)
}
func (c content) Equals(o content) bool {
	return c == o
}

func TestDiff(t *testing.T) {
	diff := make(differ.Differ[string, string, content])

	// initial diff -> everything is new
	out := diff.Diff("hello", "world", []content{"this is", "the first", "content"})
	if out != `> this is
> the first
> content` {
		t.Fatalf("Wrong output when diffing with the first content")
	}

	// initial store -> everything is new
	out = diff.DiffStore("hello", "world", []content{"this is", "the first", "content"})
	if out != `> this is
> the first
> content` {
		t.Fatalf("Wrong output when storing first content")
	}

	// same content -> nothing is new
	out = diff.DiffStore("hello", "world", []content{"this is", "the first", "content"})
	if out != "" {
		t.Fatalf("Wrong output when checking for zero diff of first content")
	}

	// changed content -> something, not everything is new
	out = diff.DiffStore("hello", "world", []content{"this is", "the second", "content"})
	if out != `< the first
> the second` && out != `> the second
< the first
` {
		t.Fatalf("Wrong output when changing the content. '%s'", out)
	}

	// initial diff with different path-> everything is new
	out = diff.Diff("hello", "test", []content{"this is", "the first", "test with different path"})
	if out != `> this is
> the first
> test with different path` {
		t.Fatalf("Wrong output when diffing with other path")
	}
}
