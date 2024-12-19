package cmdsplit_test

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
	cmdsplit "signalbot_go/internal/cmdSplit"
	"testing"
	"reflect"
)

func TestSplitSucc(t *testing.T) {
	strL, err := cmdsplit.Split(`hello world "arg with\ space" space\ arg backslash\\ 'arg with\ space'`)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}

	if len(strL) != 6 {
		t.Fatalf("Wrong amount of elements. Is %d but should be 6", len(strL))
	}
	if strL[0] != "hello" {
		t.Fatalf("Was: %v but should be %v", strL[0], "hello")
	}
	if strL[1] != "world" {
		t.Fatalf("Was: %v but should be %v", strL[1], "world")
	}
	if strL[2] != "arg with space" {
		t.Fatalf("Was: %v but should be %v", strL[2], "arg with space")
	}
	if strL[3] != "space arg" {
		t.Fatalf("Was: %v but should be %v", strL[3], "space arg")
	}
	if strL[4] != "backslash\\" {
		t.Fatalf("Was: %v but should be %v", strL[4], "backslash\\")
	}
	if strL[5] != "arg with space" {
		t.Fatalf("Was: %v but should be %v", strL[5], "arg with space")
	}
}

func TestSplitFail(t *testing.T) {
	_, err := cmdsplit.Split(`hello world "arg with`)
	if err != cmdsplit.ErrTrailingQuoteEscape {
		t.Fatalf("Should have returned %v", cmdsplit.ErrTrailingQuoteEscape)
	}
}

func TestUnescapeSucc(t *testing.T) {
	str, err := cmdsplit.Unescape(`hello\ world\\slash "nothing useful"`)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}

	if str != `hello world\slash "nothing useful"` {
		t.Fatalf("Was: %v but should be %v", str, `hello world\slash "nothing useful"`)
	}

	str, err = cmdsplit.Unescape(`hello\ world\\slash "nothing useful`)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}

	if str != `hello world\slash "nothing useful` {
		t.Fatalf("Was: %v but should be %v", str, `hello world\slash "nothing useful`)
	}
}

func TestUnescapeFail(t *testing.T) {
	_, err := cmdsplit.Unescape(`hello\ world\`)
	if err != cmdsplit.ErrTrailingEscape {
		t.Fatalf("Should have returned %v", cmdsplit.ErrTrailingEscape)
	}
}

func TestParseNumberRange(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
		wantErr  bool
	}{
		// Valid cases
		{"1-5", []int{1, 2, 3, 4, 5}, false},
		{"1-5,10-15", []int{1, 2, 3, 4, 5, 10, 11, 12, 13, 14, 15}, false},
		{"1-5,10,15-20,25", []int{1, 2, 3, 4, 5, 10, 15, 16, 17, 18, 19, 20, 25}, false},
		{"1,2,3", []int{1, 2, 3}, false},
		{"2,3", []int{2, 3}, false},

		// Invalid cases
		{"1-", nil, true},
		{"10-5", nil, true},
		{"1-a", nil, true},
		{"", nil, true},

		// Edge cases
		{"1-1", []int{1}, false}, // Single element range
		{"1-2,2-3", []int{1, 2, 2, 3}, false}, // Overlapping ranges
		{"1,1,2,2-3", []int{1, 1, 2, 2, 3}, false}, // Duplicate numbers
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := cmdsplit.ParseNumberRange(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNumberRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseNumberRange() = %v, want %v", got, tt.expected)
			}
		})
	}
}
