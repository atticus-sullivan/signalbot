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
	signaldbus "signalbot_go/signalcli/drivers/dbus"
)

// configuration of a signalServer. Can be parsed by yaml
// TODO note on concurrency
type SignalServerCfg struct {
	Dbus           signaldbus.DbusType   `yaml:"dbus"`
	PortSendMsg    uint16                `yaml:"portSendMsg"`
	PortVirtRcvMsg uint16                `yaml:"portVirtRcvMsg"`
	Handlers       map[string]HandlerCfg `yaml:"handlers"` // maps name to prefix

	// just to have a place where to define anchors to alias to laster
	Chats []string `yaml:"chats"`
	Users []string `yaml:"users"`
}

// check if stored values are valid
func (c *SignalServerCfg) Validate() error {
	if c.Dbus != signaldbus.SystemBus && c.Dbus != signaldbus.SessionBus {
		return fmt.Errorf("Invalid dbus type")
	}
	for _, h := range c.Handlers {
		if err := h.Validate(); err != nil {
			return err
		}
	}
	// no validation of Chats and Users as it is only for anchors in the config
	return nil
}
