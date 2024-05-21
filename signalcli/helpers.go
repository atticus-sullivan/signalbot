package signalcli

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

// send message to arbitrary recipient.If groupID is empty, send to
// recipient. If groupID is set, the message is sent to the group (and the
// recipient is ignored)
func (s *Account) SendGeneric(message string, attachments []string, recipient string, groupID []byte, notify bool) (timestamp int64, err error) {
	if len(groupID) > 0 {
		// send group message ignoring recipient
		return s.driver.SendGroupMessage(message, attachments, groupID)
	} else {
		// send normal personal message
		return s.driver.SendMessage(message, attachments, recipient, notify)
	}
}

// respond to a certain message. The recipient/groupID will be extracted from the message
func (s *Account) Respond(message string, attachments []string, m *Message, notify bool) (timestamp int64, err error) {
	dst := m.Sender
	if m.Sender == s.SelfNr {
		dst = m.Receiver
	}
	return s.SendGeneric(message, attachments, dst, m.GroupId, notify)
}

func (s *Account) GetGroupName(groupId []byte) (string, error) {
	return s.driver.GetGroupName(groupId)
}
