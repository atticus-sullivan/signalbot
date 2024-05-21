package signalsender

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

import "signalbot_go/signalcli"

// an interface which allows to send data to signal
type SignalSender interface {
	// send message to arbitrary recipient.If groupID is empty, send to
	// recipient. If groupID is set, the message is sent to the group (and the
	// recipient is ignored)
	SendGeneric(message string, attachments []string, recipient string, groupID []byte, notify bool) (timestamp int64, err error)
	// respond to a certain message. The recipient/groupID will be extracted from the message
	Respond(message string, attachments []string, m *signalcli.Message, notify bool) (timestamp int64, err error)
}
