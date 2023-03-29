package signaldbus

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/godbus/dbus/v5"
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

type ReceiptHandler interface {
	handle(*Receipt)
}
type ReceiptHandlerFunc func(*Receipt)

func (f ReceiptHandlerFunc) handle(m *Receipt) {
	f(m)
}
func (s *Account) AddReceiptHandler(handler ReceiptHandler) {
	s.receiptHandlersChann <- &handler
}
func (s *Account) AddReceiptHandlerFunc(handler func(*Receipt)) error {
	if handler == nil {
		return fmt.Errorf("signal-cli: trying to acc a nil receipt handler func")
	}
	s.AddReceiptHandler(ReceiptHandlerFunc(handler))
	return nil
}

// The sync message is received when the user sends a message from a linked
// device.
type SyncMessage struct {
	Message
	// DBus code for destination
	Destination string
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
	Timestamp int64
	// Phone number of the sender
	Sender string
}

// This signal is received whenever we get a private message or a message is
// posted in a group we are an active member.
type Message struct {
	// Integer value that is used by the system to send a ReceiptReceived reply
	Timestamp int64
	// Phone number of the sender
	Sender string
	// Byte array representing the internal group identifier (empty when
	// private message)
	GroupId []byte
	// Message text
	Message string
	// String array of filenames in the signal-cli storage
	// (~/.local/share/signal-cli/attachments/)
	Attachments []string
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

func NewSyncMessage(v *dbus.Signal) *SyncMessage {
	msg := SyncMessage{
		Message: Message{
			Timestamp:   v.Body[0].(int64),
			Sender:      v.Body[1].(string),
			GroupId:     v.Body[3].([]byte),
			Message:     v.Body[4].(string),
			Attachments: v.Body[5].([]string),
		},
		Destination: v.Body[2].(string),
	}
	return &msg
}

func NewReceipt(v *dbus.Signal) *Receipt {
	msg := Receipt{
		Timestamp: v.Body[0].(int64),
		Sender:    v.Body[1].(string),
	}
	return &msg
}

func NewMessage(v *dbus.Signal) *Message {
	msg := Message{
		Timestamp:   v.Body[0].(int64),
		Sender:      v.Body[1].(string),
		GroupId:     v.Body[2].([]byte),
		Message:     v.Body[3].(string),
		Attachments: v.Body[4].([]string),
	}
	return &msg
}

func NewMessageFromReader(r io.Reader) (*Message, error) {
	rb := bufio.NewReader(r)
	m := Message{}

	gidB, err := rb.ReadBytes('\n')
	if err != nil && err == io.EOF {
		return nil, err
	}
	gidS := strings.TrimSpace(string(gidB))
	gid, err := hex.DecodeString(gidS)
	if err != nil && err == io.EOF {
		return nil, err
	}

	source, err := rb.ReadBytes('\n')
	if err != nil && err == io.EOF {
		return nil, err
	}

	msg, err := rb.ReadBytes('\n')
	if err != nil && err == io.EOF {
		return nil, err
	}

	m.Message = strings.TrimSpace(string(msg))
	if len(gidS) != 0 {
		m.GroupId = gid
	}
	if len(source) != 0 {
		m.Sender = strings.TrimSpace(string(source))
	}
	return &m, nil
}
