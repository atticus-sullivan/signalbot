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
	"errors"
	"fmt"
	"signalbot_go/internal/act"
)

var ErrInvalidUser error = errors.New("Invalid User specified")
var ErrInvalidChat error = errors.New("Invalid Chat specified")
var ErrInvalidACTDepth error = errors.New("ACT has the wrong depth")

type Accesscontrol act.ACT

// Validate if stored information is valid
func (a *Accesscontrol) Validate() error {
	if err := a.Default.Validate(); err != nil {
		return err
	}

	for user, actA := range a.Children {
		if !validPhoneNr(user) {
			return ErrInvalidUser
		}

		if err := actA.Default.Validate(); err != nil {
			return err
		}
		for chat, actB := range actA.Children {
			if !validChat(chat) {
				return ErrInvalidChat
			}

			if err := actB.Default.Validate(); err != nil {
				return err
			}
			if len(actB.Children) != 0 {
				return ErrInvalidACTDepth
			}
		}
	}
	return nil
}

func (a *Accesscontrol) Check(user string, chat string) error {
	actA, set := a.Children[user]
	if !set {
		if a.Default.Blocked() {
			return fmt.Errorf("Not allowed. User: %s not set, default: %s", user, a.Default)
		}
		if a.Default.Allowed() {
			return nil
		}
		// unset
		return fmt.Errorf("Not allowed. User: %s not set, default unset", user)
	}
	actB, set := actA.Children[chat]
	if !set {
		def := actA.Default
		if actA.Default.Unset() {
			def = a.Default
		}
		if def.Blocked() {
			return fmt.Errorf("Not allowed. User: %s Chat: %s not set, default: %s", user, chat, def)
		}
		if def.Allowed() {
			return nil
		}
	}
	def := actB.Default
	if actB.Default.Unset() {
		def = actA.Default
	}
	if def.Blocked() {
		return fmt.Errorf("Not allowed. User: %s Chat: %s set but disallowed", user, chat)
	}
	if def.Allowed() {
		return nil
	}
	return fmt.Errorf("Invalid ACT")
}
