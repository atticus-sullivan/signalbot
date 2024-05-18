package act_test

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
	"signalbot_go/internal/act"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTestUnmarshalYAML(t *testing.T) {
	r := strings.NewReader(`
default: "Block"
children:
  C1A:
    default: "Allow"
    children:
      C2A: "Block"
  C1B:
    default: "Block"
    children:
      C2A: "Allow"
`)
	dec := yaml.NewDecoder(r)
	var out act.ACT
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("Err: %v", err)
	}

	if !out.Default.Blocked() {
		t.Fatalf("root default is not set to blocked")
	}
	if len(out.Children) != 2 {
		t.Fatalf("root has not the right amount of children")
	}

	c, ok := out.Children["C1A"]
	if !ok {
		t.Fatalf("root has not the child 'C1A'")
	}
	if !c.Default.Allowed() {
		t.Fatalf("'C1A' default is not set to allowed")
	}
	if len(c.Children) != 1 {
		t.Fatalf("'C1A' has not the right amount of children")
	}

	c, ok = c.Children["C2A"]
	if !ok {
		t.Fatalf("'C1A' has not the child 'C2A'")
	}
	if !c.Default.Blocked() {
		t.Fatalf("'C2A' default is not set to allowed")
	}
	if len(c.Children) != 0 {
		t.Fatalf("'C2A' has not the right amount of children")
	}

	c, ok = out.Children["C1B"]
	if !ok {
		t.Fatalf("root has not the child 'C1B'")
	}
	if !c.Default.Blocked() {
		t.Fatalf("'C1B' default is not set to blocked")
	}
	if len(c.Children) != 1 {
		t.Fatalf("'C1B' has not the right amount of children")
	}

	c, ok = c.Children["C2A"]
	if !ok {
		t.Fatalf("'C1B' has not the child 'C2A'")
	}
	if !c.Default.Allowed() {
		t.Fatalf("'C2A' default is not set to allowed")
	}
	if len(c.Children) != 0 {
		t.Fatalf("'C2A' has not the right amount of children")
	}
}
