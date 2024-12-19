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

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

// ///////////
// SIGNALS //
// ///////////

type MessageHandler interface {
	handle(*Message)
}
type MessageHandlerFunc func(*Message)

func (f MessageHandlerFunc) handle(m *Message) {
	f(m)
}

func (s *Account) AddMessageHandler(handler MessageHandler) {
	s.messageHandlersChann <- &handler
}
func (s *Account) AddMessageHandlerFunc(handler func(*Message)) error {
	if handler == nil {
		return fmt.Errorf("signal-cli: trying to acc a nil message handler func")
	}
	s.AddMessageHandler(MessageHandlerFunc(handler))
	return nil
}


type SyncMessageHandler interface {
	handle(*SyncMessage)
}
type SyncMessageHandlerFunc func(*SyncMessage)

func (f SyncMessageHandlerFunc) handle(m *SyncMessage) {
	f(m)
}
func (s *Account) AddSyncMessageHandler(handler SyncMessageHandler) {
	s.syncMessageHandlersChann <- &handler
}
func (s *Account) AddSyncMessageHandlerFunc(handler func(*SyncMessage)) error {
	if handler == nil {
		return fmt.Errorf("signal-cli: trying to acc a nil sync message handler func")
	}
	s.AddSyncMessageHandler(SyncMessageHandlerFunc(handler))
	return nil
}


// The sync message is received when the user sends a message from a linked
// device.
type SyncMessage struct {
	Message `yaml:",inline"`
	// Phonenumber of the destination
	Destination string `yaml:"dst"`
}

func (m *SyncMessage) String() string {
	builder := strings.Builder{}
	builder.WriteRune('{')
	builder.WriteString("Dst: ")
	builder.WriteString(m.Destination)
	builder.WriteRune(' ')
	builder.WriteString("Msg: ")
	builder.WriteString(m.Message.String())
	builder.WriteRune('}')
	return builder.String()
}

// This signal is sent by each recipient (e.g. each group member) after the
// message was successfully delivered to the device.
type Receipt struct {
	// Integer value that can be used to associate this e.g. with a
	// sendMessage()
	Timestamp int64 `yaml:"ts"`
	// Phone number of the sender
	Sender string `yaml:"sender"`
}

// This signal is received whenever we get a private message or a message is
// posted in a group we are an active member.
type Message struct {
	// Integer value that is used by the system to send a ReceiptReceived reply
	Timestamp int64 `yaml:"ts"`
	// Phone number of the sender
	Sender string `yaml:"sender"`
	// Phone number of the reveicer
	Receiver string `yaml:"receiver"`
	// Byte array representing the internal group identifier (empty when
	// private message)
	GroupId []byte `yaml:"gid,flow"`
	// Either the hex representation of the chat or the phonenumber identifying
	// the chat (usually the sender, or on sync messages the receiver)
	Chat string `yaml:"chat"`
	// Message text
	Message string `yaml:"msg"`
	// String array of filenames in the signal-cli storage
	// (~/.local/share/signal-cli/attachments/)
	Attachments []string `yaml:"att,flow"`
}

func (m *Message) String() string {
	builder := strings.Builder{}
	builder.WriteRune('{')
	builder.WriteString("TS: ")
	builder.WriteString(fmt.Sprintf("%d", m.Timestamp))
	builder.WriteRune(' ')
	builder.WriteString("Sender: ")
	builder.WriteString(m.Sender)
	builder.WriteRune(' ')
	builder.WriteString("Receiver: ")
	builder.WriteString(m.Receiver)
	builder.WriteRune(' ')
	builder.WriteString("GID: ")
	builder.WriteString(hex.EncodeToString(m.GroupId))
	builder.WriteRune(' ')
	builder.WriteString("Msg: ")
	builder.WriteString(m.Message)
	builder.WriteRune(' ')
	builder.WriteString("Att: ")
	builder.WriteString(fmt.Sprintf("%v", m.Attachments))
	builder.WriteRune('}')
	return builder.String()
}

func NewMessageFromReader(r io.Reader, self string) (*Message, error) {
	rb := bufio.NewReader(r)

	gidB, err := rb.ReadBytes('\n')
	if err != nil && err == io.EOF {
		return nil, err
	}
	gidS := strings.TrimSpace(string(gidB))
	gid, err := hex.DecodeString(gidS)
	if err != nil && err == io.EOF {
		return nil, err
	}

	sender, err := rb.ReadBytes('\n')
	if err != nil && err == io.EOF {
		return nil, err
	}

	receiver, err := rb.ReadBytes('\n')
	if err != nil && err == io.EOF {
		return nil, err
	}

	msg, err := rb.ReadBytes('\n')
	if err != nil && err == io.EOF {
		return nil, err
	}

	m := Message{}

	m.Message = strings.TrimSpace(string(msg))

	if len(gidS) != 0 {
		m.GroupId = gid
	}

	if len(sender) != 0 {
		m.Sender = strings.TrimSpace(string(sender))
	}

	if len(receiver) != 0 {
		m.Receiver = strings.TrimSpace(string(receiver))
	}

	// fill chat
	if len(m.GroupId) > 0 {
		m.Chat = hex.EncodeToString(m.GroupId)
	}
	if m.Sender == self {
		m.Chat = m.Receiver
	}
	m.Chat = m.Sender

	return &m, nil
}
