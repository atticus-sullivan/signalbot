package signalconsole

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
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"signalbot_go/signalcli"
)

type SignalCliDriver struct {
	driverInter signalcli.InterDriverToAcc
	selfNr      string
	log         *slog.Logger
	ctx     context.Context
	cFunc       context.CancelFunc
}

const SELF_NR = "+4900"

func NewSignalJsonRpcDriver(log *slog.Logger, unixSocket string, selfNr string) (*SignalCliDriver, error) {
	scd := &SignalCliDriver{
		selfNr: SELF_NR,
		log:    log,
	}
	scd.ctx, scd.cFunc = context.WithCancel(context.Background())
	return scd, nil
}

func (scd *SignalCliDriver) GetSelfNumber() (string, error) {
	return scd.selfNr, nil
}

var PROMPT = fmt.Sprintf("%s: ", SELF_NR)

func (scd *SignalCliDriver) Start() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(PROMPT)
		msg, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		m := signalcli.Message{
			Timestamp:   0,
			Sender:      SELF_NR,
			Receiver:    SELF_NR,
			Message:     msg,
		}
		m.Chat = m.Sender
		scd.driverInter.MessageChan <- &m
		select {
		case <- scd.ctx.Done():
			return
		default:
		}
	}
}

func (scd *SignalCliDriver) Close() {
	scd.cFunc()
}

func (scd *SignalCliDriver) SendMessage(message string, attachments []string, recipient string, notifySelf bool) (int64, error) {
	fmt.Print("\x33[2K\r")
	prefix := fmt.Sprintf("> %s:", recipient)
	message = strings.ReplaceAll(message, "\n", "\n"+strings.Repeat(" ", len(prefix)+1))
	fmt.Println(prefix, message)
	fmt.Print(PROMPT)
	return 0, nil
}

func (scd *SignalCliDriver) SendGroupMessage(message string, attachments []string, groupId []byte) (int64, error) {
	gn, _ := scd.GetGroupName(groupId)
	fmt.Println("\332K\331G", "=> ", gn, ": ", message)
	fmt.Print(PROMPT)
	return 0, nil
}

func (scd *SignalCliDriver) GetGroupName(groupId []byte) (string, error) {
	return hex.EncodeToString(groupId), nil
}

func (scd *SignalCliDriver) SetInterface(inter signalcli.InterDriverToAcc) (err error) {
	scd.driverInter = inter
	return nil
}
