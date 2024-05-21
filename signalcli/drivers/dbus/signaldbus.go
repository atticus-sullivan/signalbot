package signaldbus

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
	"log/slog"
	"strings"

	"signalbot_go/signalcli"

	"github.com/godbus/dbus/v5"
)

// enum for different busses to listen on
type DbusType string

const (
	SessionBus = "sessionBus"
	SystemBus  = "systemBus"
)

type SignalCliDriver struct {
	conn *dbus.Conn
	obj  dbus.BusObject
	driverInter signalcli.InterDriverToAcc
	signals <-chan *dbus.Signal
	selfNr string
	log *slog.Logger
	stop                     chan struct{}
}

func NewSignalDbusDriver(log *slog.Logger, busType DbusType) (*SignalCliDriver, error) {
	var err error

	ret := SignalCliDriver{
		log: log,
		stop:                     make(chan struct{}),
	}

	switch busType {
	case SessionBus:
		ret.conn, err = dbus.ConnectSessionBus()
	case SystemBus:
		ret.conn, err = dbus.ConnectSystemBus()
	default:
		return nil, fmt.Errorf("signal-cli: wrong busType\n")
	}
	if err != nil {
		return nil, err
	}
	ret.obj = ret.conn.Object("org.asamk.Signal", "/org/asamk/Signal")

	ret.selfNr, err = ret.GetSelfNumber()


	return &ret, nil
}

func (d *SignalCliDriver) SetInterface(inter signalcli.InterDriverToAcc) (err error) {
	d.driverInter = inter

	if err = d.conn.AddMatchSignal(
		// TODO maybe only add signals for which to listen to here
		dbus.WithMatchInterface("org.asamk.Signal"),
	); err != nil {
		return err
	}
	signals := make(chan *dbus.Signal, 20)
	d.signals = signals
	d.conn.Signal(signals)
	return nil
}

func (d* SignalCliDriver) Start() {

	running := true
	for running {
		select {
		case <-d.stop:
			running = false
		case ele := <-d.signals:
			if ele == nil {
				continue
			}
			d.log.Debug(fmt.Sprintf("%v", ele))
			switch ele.Name {
			case "org.asamk.Signal.SyncMessageReceived":
				msg := NewSyncMessage(ele, d.selfNr)
				d.log.Debug("driver syncMsg", "msg", msg, "chan", d.driverInter.SyncMessageChan)
				d.driverInter.SyncMessageChan <- msg
			case "org.asamk.Signal.MessageReceived":
				msg := NewMessage(ele, d.selfNr)
				d.log.Debug("driver Msg", "msg", msg, "chan", d.driverInter.MessageChan)
				d.driverInter.MessageChan <- msg

			// known, but currently not used
			case "org.asamk.Signal.ReceiptReceived":
			case "org.asamk.Signal.SyncMessageReceivedV2":
			case "org.asamk.Signal.ReceiptReceivedV2":
			case "org.samk.Signal.MessageReceivedV2":
			default:
				d.log.Info("Unknown signal caught: ", ele.Name, ele)
			}
		}
	}
}

func (d* SignalCliDriver) Close() {
	d.stop <- struct{}{}
	d.conn.Close()
}

func NewSyncMessage(v *dbus.Signal, self string) *signalcli.SyncMessage {
	msg := signalcli.SyncMessage{
		Message: signalcli.Message{
			Timestamp:   v.Body[0].(int64),
			Sender:      v.Body[1].(string),
			GroupId:     v.Body[3].([]byte),
			Message:     v.Body[4].(string),
			Attachments: v.Body[5].([]string),
		},
		Destination: v.Body[2].(string),
	}
	msg.Message.Receiver = msg.Destination

	// fill chat
	if len(msg.GroupId) > 0 {
		msg.Chat = hex.EncodeToString(msg.GroupId)
	} else {
		if msg.Sender == self {
			msg.Chat = msg.Receiver
		}
		msg.Chat = msg.Sender
	}

	return &msg
}

// func NewReceipt(v *dbus.Signal, self string) *signalcli.Receipt {
// 	msg := signalcli.Receipt{
// 		Timestamp: v.Body[0].(int64),
// 		Sender:    v.Body[1].(string),
// 	}
// 	return &msg
// }

func NewMessage(v *dbus.Signal, self string) *signalcli.Message {
	msg := signalcli.Message{
		Timestamp:   v.Body[0].(int64),
		Sender:      v.Body[1].(string),
		GroupId:     v.Body[2].([]byte),
		Message:     v.Body[3].(string),
		Attachments: v.Body[4].([]string),
	}
	msg.Receiver = self

	// fill chat
	if len(msg.GroupId) > 0 {
		msg.Chat = hex.EncodeToString(msg.GroupId)
	} else {
		if msg.Sender == self {
			msg.Chat = msg.Receiver
		}
		msg.Chat = msg.Sender
	}

	return &msg
}

func NewMessageFromReader(r io.Reader, self string) (*signalcli.Message, error) {
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

	m := signalcli.Message{}

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
