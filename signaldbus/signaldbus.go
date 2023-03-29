package signaldbus

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"golang.org/x/exp/slog"
)

// enum for different busses to listen on
type DbusType string

const (
	SessionBus = "sessionBus"
	SystemBus  = "systemBus"
)

// an Interface to the signal-cli dbus -- signel account mode
// Needs to be closed.
// Is not safe for concurrency because of the obj member which is not safe for
// concurrency (might add a layer of abstraction which does mutexes or obtain a
// new obj for each call)
// And if signalsListening is safe for concurrency is unsure
// Create new objects with NewAccount
type Account struct {
	signalsListening bool

	messageHandlersChann     chan *MessageHandler
	syncMessageHandlersChann chan *SyncMessageHandler
	receiptHandlersChann     chan *ReceiptHandler
	stop                     chan interface{}

	conn *dbus.Conn
	obj  dbus.BusObject

	selfNr string

	signals <-chan *dbus.Signal

	log *slog.Logger
}

// close the account. Stops listening for signals and closes the connection to
// signal.
// Might block if ListenForSignals was never called!
func (s *Account) Close() {
	s.stop <- true
	s.conn.Close()
}

// create a new Account object.
func NewAccount(log *slog.Logger, busType DbusType) (acc *Account, err error) {
	acc = &Account{
		log:                      log,
		signalsListening:         false,
		messageHandlersChann:     make(chan *MessageHandler, 5),
		syncMessageHandlersChann: make(chan *SyncMessageHandler, 5),
		receiptHandlersChann:     make(chan *ReceiptHandler, 5),
		stop:                     make(chan interface{}),
	}

	switch busType {
	case SessionBus:
		acc.conn, err = dbus.ConnectSessionBus()
	case SystemBus:
		acc.conn, err = dbus.ConnectSystemBus()
	default:
		return nil, fmt.Errorf("signal-cli: wrong busType\n")
	}
	if err != nil {
		return nil, err
	}
	acc.obj = acc.conn.Object("org.asamk.Signal", "/org/asamk/Signal")

	acc.selfNr, err = acc.GetSelfNumber()
	if err != nil {
		return nil, err
	}

	if err = acc.conn.AddMatchSignal(
		// TODO maybe only add signals for which to listen to here
		dbus.WithMatchInterface("org.asamk.Signal"),
	); err != nil {
		return nil, err
	}
	// scope this to avoid accidentally working sending to the channel
	{
		signals := make(chan *dbus.Signal, 20)
		acc.signals = signals
		acc.conn.Signal(signals)
	}

	return acc, nil
}

// starts listening for signals from signal-cli. Waits until Listening is
// completely set up
func (s *Account) ListenForSignals() {
	sync := make(chan int)
	go s.ListenForSignalsWithSync(sync)
	<-sync
}

// starts listening for signals from signal-cli. The sync chan will be send to
// after setting up listening is finished.
func (s *Account) ListenForSignalsWithSync(sync chan<- int) {
	if s.signalsListening {
		s.log.Warn("signals already connected. Skipping this call")
		sync <- 1
		return
	}
	s.signalsListening = true

	messageHandlers := []*MessageHandler{}
	syncMessageHandlers := []*SyncMessageHandler{}
	receiptMessageHandlers := []*ReceiptHandler{}

	var (
		hm *MessageHandler
		hs *SyncMessageHandler
		hr *ReceiptHandler
	)

	s.log.Info("signal-cli: listening")
	sync <- 1
	running := true
	for running {
		select {
		case <-s.stop:
			running = false
		case hs = <-s.syncMessageHandlersChann:
			syncMessageHandlers = append(syncMessageHandlers, hs)
		case hm = <-s.messageHandlersChann:
			messageHandlers = append(messageHandlers, hm)
		case hr = <-s.receiptHandlersChann:
			receiptMessageHandlers = append(receiptMessageHandlers, hr)
		case ele := <-s.signals:
			if ele == nil {
				continue
			}
			// s.log.Info(fmt.Sprintf("%v", ele))
			switch ele.Name {
			case "org.asamk.Signal.SyncMessageReceived":
				msg := NewSyncMessage(ele)
				s.log.Info("Sync", slog.Any("body", msg.String()))
				for _, h := range syncMessageHandlers {
					(*h).handle(msg)
				}
			case "org.asamk.Signal.ReceiptReceived":
				s.log.Info("Receipt", slog.Any("body", ele.Body))
				msg := NewReceipt(ele)
				for _, h := range receiptMessageHandlers {
					(*h).handle(msg)
				}
			case "org.asamk.Signal.MessageReceived":
				msg := NewMessage(ele)
				s.log.Info("Message", slog.Any("body", msg.String()))
				for _, h := range messageHandlers {
					(*h).handle(msg)
				}
			case "org.asamk.Signal.SyncMessageReceivedV2":
			case "org.asamk.Signal.ReceiptReceivedV2":
			case "org.samk.Signal.MessageReceivedV2":
			default:
				s.log.Info("Unknown signal caught: ", ele.Name, ele)
			}
		}
	}
}

// TODO Signal.Control interface
// TODO Signal.Group interface
// TODO Signal.Device interface
// TODO Signal.Configuration interface
