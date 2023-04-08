package act

import (
	"errors"

	"gopkg.in/yaml.v3"
)

// stores if the access is granted or blocked
type Capability string

const (
	allow Capability = "Allow"
	block Capability = "Block"
	unset Capability = ""
)

func (c Capability) Blocked() bool {
	return c == block
}
func (c Capability) Allowed() bool {
	return c == allow
}
func (c Capability) Unset() bool {
	return c == unset
}

var ErrInvalidCap error = errors.New("Invalid Capability value")

func (c Capability) Validate() error {
	if c != allow && c != block && c != unset {
		return ErrInvalidCap
	}
	return nil
}

// Capability is a stringer
func (c Capability) String() string {
	return string(c)
}

type ACT struct {
	Default Capability `yaml:"default"`
	Children map[string]ACT `yaml:"children"`
}

func (a *ACT) UnmarshalYAML(value *yaml.Node) error {
	var capability Capability
	// try decoding string
	if err := value.Decode(&capability); err == nil {
		// string was provided
		a.Default = capability
		return nil
	}
	// try decoding ACT
	type rawACT ACT // new type to avoid endless recursion
	if err := value.Decode((*rawACT)(a)); err != nil {
		return err
	}
	// ACT already got populated
	return nil
}
