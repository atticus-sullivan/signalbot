package signalserver

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
	"fmt"
	"signalbot_go/internal/signalsender"
	"signalbot_go/signaldbus"
	"strings"
)

// can handle a signal-message
type Handler interface {
	Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message))
	Start(virtRcv func(*signaldbus.Message)) error
	Close(virtRcv func(*signaldbus.Message))
}

// config for a handler. Can be parsed from yaml
// TODO note on concurrency
type HandlerCfg struct {
	Prefixes []string      `yaml:"prefixes"`
	Help     string        `yaml:"help"`
	Access   Accesscontrol `yaml:"access"`
}

// validate the stored data
func (c *HandlerCfg) Validate() error {
	for _, p := range c.Prefixes {
		if strings.ContainsRune(p, ' ') {
			return fmt.Errorf("Invalid prefix: %v (cannot contain space)", p)
		}
	}
	return c.Access.Validate()
}
