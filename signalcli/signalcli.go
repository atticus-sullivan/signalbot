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
	"log/slog"
)

type Driver interface {
	SendMessage(message string, attachments []string, recipient string, notifySelf bool) (timestamp int64, err error)
	SendGroupMessage(message string, attachments []string, groupId []byte) (timestamp int64, err error)
	GetGroupName(groupId []byte) (string, error)
	GetSelfNumber() (number string, err error)
	SetInterface(inter InterDriverToAcc) (err error)
	Start()
	Close()
}

type InterAccToDriver struct {
	MessageChan <-chan *Message
	SyncMessageChan <-chan *SyncMessage
}

type InterDriverToAcc struct {
	MessageChan chan<- *Message
	SyncMessageChan chan<- *SyncMessage
}

// an Interface to the signal-cli dbus -- signle account mode
// Needs to be closed.
// Is not safe for concurrency because of the obj member which is not safe for
// concurrency (might add a layer of abstraction which does mutexes or obtain a
// new obj for each call)
// And if signalsListening is safe for concurrency is unsure
// Create new objects with NewAccount
type Account struct {
	driver Driver

	signalsListening bool

	messageHandlersChann     chan *MessageHandler
	syncMessageHandlersChann chan *SyncMessageHandler
	stop                     chan interface{}

	SelfNr string

	driverInter InterAccToDriver

	log *slog.Logger
}

// close the account. Stops listening for signals and closes the connection to
// signal.
// Might block if ListenForSignals was never called!
func (s *Account) Close() {
	s.stop <- true
	s.driver.Close()
}

// create a new Account object.
func NewAccount(log *slog.Logger, c Driver) (acc *Account, err error) {
	msgChan := make(chan *Message, 5)
	syncMsgChan := make(chan *SyncMessage, 5)

	acc = &Account{
		driver: c,

		signalsListening:         false,

		messageHandlersChann:     make(chan *MessageHandler, 5),
		syncMessageHandlersChann: make(chan *SyncMessageHandler, 5),
		stop:                     make(chan interface{}),

		driverInter: InterAccToDriver{
			MessageChan: msgChan,
			SyncMessageChan: syncMsgChan,
		},

		log:                      log,
	}

	acc.SelfNr, err = acc.driver.GetSelfNumber()
	if err != nil {
		return nil, err
	}

	i := InterDriverToAcc{
		MessageChan: msgChan,
		SyncMessageChan: syncMsgChan,
	}
	acc.log.Debug("init channels", "driver", i.MessageChan, "acc", acc.driverInter.MessageChan)
	acc.driver.SetInterface(i)

	return acc, nil
}

// starts listening for signals from signal-cli. Waits until Listening is
// completely set up
func (s *Account) ListenForSignals() {
	sync := make(chan struct{})
	go s.ListenForSignalsWithSync(sync)
	<-sync
}

// starts listening for signals from signal-cli. The sync chan will be send to
// after setting up listening is finished.
func (s *Account) ListenForSignalsWithSync(sync chan<- struct{}) {
	if s.signalsListening {
		s.log.Warn("signals already connected. Skipping this call")
		sync <- struct{}{}
		return
	}
	s.signalsListening = true

	messageHandlers := []*MessageHandler{}
	syncMessageHandlers := []*SyncMessageHandler{}

	var (
		hm *MessageHandler
		hs *SyncMessageHandler
	)

	go s.driver.Start()

	s.log.Info("signal-cli: listening")
	sync <- struct{}{}
	running := true
	for running {
		select {
		case <-s.stop:
			running = false

		case hs = <-s.syncMessageHandlersChann:
			s.log.Info("syncMessage handler registered")
			syncMessageHandlers = append(syncMessageHandlers, hs)
		case hm = <-s.messageHandlersChann:
			s.log.Info("Message handler registered")
			messageHandlers = append(messageHandlers, hm)

		case ele := <-s.driverInter.MessageChan:
			s.log.Info("message from driver", "msg", ele)
			if ele == nil {
				continue
			}
			s.log.Info("Message", "body", ele.String())
			for _, h := range messageHandlers {
				(*h).handle(ele)
			}
		case ele := <-s.driverInter.SyncMessageChan:
			s.log.Info("message from driver", "msg", ele)
			if ele == nil {
				continue
			}
			s.log.Info("Sync", "body", ele.String())
			for _, h := range syncMessageHandlers {
				(*h).handle(ele)
			}

		}
	}
}

// TODO Signal.Control interface
// TODO Signal.Group interface
// TODO Signal.Device interface
// TODO Signal.Configuration interface
