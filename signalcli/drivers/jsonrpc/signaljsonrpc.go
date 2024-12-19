package signaljsonrpc

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
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"strings"

	"signalbot_go/signalcli"

	"golang.org/x/exp/jsonrpc2"
)

var (
	ErrMsgUnset = errors.New("Message unset")
)

type SignalCliDriver struct {
	driverInter signalcli.InterDriverToAcc
	selfNr string
	log *slog.Logger
	conn *jsonrpc2.Connection
}

func NewSignalJsonRpcDriver(log *slog.Logger, unixSocket string, selfNr string) (*SignalCliDriver, error) {
	var err error

	ret := SignalCliDriver{
		log: log,
		selfNr: selfNr,
	}

	ret.conn, err = jsonrpc2.Dial(
		context.Background(),
		jsonrpc2.NetDialer("unix", unixSocket, net.Dialer{}),
		jsonrpc2.ConnectionOptions{
			Framer: RawFramerNewline(),
			Handler: &ret,
		},
	)
	if err != nil {
		panic(err)
	}

	return &ret, nil
}

func (d *SignalCliDriver) GetSelfNumber() (string, error) {
	return d.selfNr, nil
}

func (d *SignalCliDriver) Handle(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
	var rcv jsonReceive
	dec := json.NewDecoder(bytes.NewReader(req.Params))
	dec.DisallowUnknownFields()

	err := dec.Decode(&rcv)
	if err != nil {
		d.log.Warn("Decoding jsonRpc message params failed", "rpc params", strings.ReplaceAll(string(req.Params), "\"", "|"))
		return nil, jsonrpc2.ErrNotHandled
	}

	if rcv.Exception != nil {
		// ignore
	} else if rcv.Envelope.TypingMessage != nil {
		// ignore
	} else if rcv.Envelope.ReceiptMessage != nil {
		// ignore
	} else if rcv.Envelope.SyncMessage != nil {
		m,err := NewSyncMessage(&rcv, d.selfNr)
		if err != nil {
			switch err {
			case ErrMsgUnset:
			default:
				d.log.Warn("Error parsing message", "err", err, "method", req.Method, "msg", strings.ReplaceAll(string(req.Params), "\"", "|"))
			}
		} else {
			d.driverInter.SyncMessageChan <- m
		}
	} else if rcv.Envelope.DataMessage != nil {
		m,err := NewDataMessage(&rcv, d.selfNr)
		if err != nil {
			switch err {
			case ErrMsgUnset:
			default:
				d.log.Warn("Error parsing message", "err", err, "method", req.Method, "msg", strings.ReplaceAll(string(req.Params), "\"", "|"))
			}
		} else {
			d.driverInter.MessageChan <- m
		}
	} else {
		d.log.Warn("Unknown message type", "rpc params", strings.ReplaceAll(string(req.Params), "\"", "|"))
	}

	return nil, jsonrpc2.ErrNotHandled
}

func (d *SignalCliDriver) SetInterface(inter signalcli.InterDriverToAcc) (err error) {
	d.driverInter = inter
	return nil
}

func (d* SignalCliDriver) Start() {}

func (d* SignalCliDriver) Close() {
	d.conn.Close()
}

func NewSyncMessage(v *jsonReceive, self string) (*signalcli.SyncMessage, error) {
	msg := signalcli.SyncMessage{
		Message: signalcli.Message{
			Timestamp:   int64(v.Envelope.SyncMessage.SentMessage.Timestamp),
			Sender:      v.Envelope.Source,
			Message:     v.Envelope.SyncMessage.SentMessage.Message,
			// Attachments: v.Envelope.SyncMessage.SentMessage.Attachments,
		},
		Destination: v.Envelope.SyncMessage.SentMessage.Destination,
	}
	if len(v.Envelope.SyncMessage.SentMessage.GroupInfo.GroupId) > 0 {
		gid,err := base64.StdEncoding.DecodeString(v.Envelope.SyncMessage.SentMessage.GroupInfo.GroupId)
		if err != nil {
			return nil, err
		}
		msg.Message.GroupId = gid
	}
	msg.Message.Receiver = msg.Destination


	if msg.Sender == "" {
		return nil, errors.New("Sender unset")
	}
	// if msg.Destination == "" {
	// 	return nil, errors.New("Destination unset")
	// }
	if msg.Message.Message == "" {
		return nil, ErrMsgUnset
	}

	// fill chat
	if len(msg.GroupId) > 0 {
		msg.Chat = hex.EncodeToString(msg.GroupId)
	} else {
		if msg.Sender == self {
			msg.Chat = msg.Receiver
		}
		msg.Chat = msg.Sender
	}

	return &msg, nil
}

func NewDataMessage(v *jsonReceive, self string) (*signalcli.Message,error) {
	msg := signalcli.Message{
		Timestamp:   int64(v.Envelope.DataMessage.Timestamp),
		Sender:      v.Envelope.Source,
		Message:     v.Envelope.DataMessage.Message,
	}
	if len(v.Envelope.DataMessage.GroupInfo.GroupId) > 0 {
		gid,err := base64.StdEncoding.DecodeString(v.Envelope.DataMessage.GroupInfo.GroupId)
		if err != nil {
			return nil, err
		}
		msg.GroupId = gid
	}
	msg.Receiver = self

	if msg.Sender == "" {
		return nil, errors.New("Sender unset")
	}
	// if msg.Destination == "" {
	// 	return nil, errors.New("Destination unset")
	// }
	if msg.Message == "" {
		return nil, ErrMsgUnset
	}

	// fill chat
	if len(msg.GroupId) > 0 {
		msg.Chat = hex.EncodeToString(msg.GroupId)
	} else {
		if msg.Sender == self {
			msg.Chat = msg.Receiver
		}
		msg.Chat = msg.Sender
	}

	return &msg, nil
}
