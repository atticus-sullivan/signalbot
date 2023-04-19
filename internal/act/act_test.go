package act_test

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
