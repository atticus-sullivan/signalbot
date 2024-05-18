package act

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
	"errors"

	"gopkg.in/yaml.v3"
)

// errors
var (
	ErrInvalidCap error = errors.New("Invalid Capability value")
)

// stores if the access is granted or blocked
type Capability string

const (
	allow Capability = "Allow"
	block Capability = "Block"
	unset Capability = ""
)

// check if capability is set to blocked
func (c Capability) Blocked() bool {
	return c == block
}

// check if capability is set to allowed
func (c Capability) Allowed() bool {
	return c == allow
}

// check if capability is set to unset
func (c Capability) Unset() bool {
	return c == unset
}

// validate if the capability is set to a valid value
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

// struct which represents an access-control-tree. Can be used e.g. as chat ->
// user -> access while having defaults for each possible subtree
type ACT struct {
	Default  Capability     `yaml:"default"`
	Children map[string]ACT `yaml:"children"`
}

// enable unmarshaling ACTs from yaml. For the leaf child only a string can be
// provided which is then used as capability
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
