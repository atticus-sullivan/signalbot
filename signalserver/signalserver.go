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

// todo: dirctly call handler for periodic stuff s.handle(m)
// todo open websocket for virtually sending messages (calls s.handle(m)). Message is passed via json (todo annotations to message struct for (un)marshalling)

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"signalbot_go/modules/buechertreff"
	"signalbot_go/modules/cmd"
	"signalbot_go/modules/fernsehserien"
	"signalbot_go/modules/freezer"
	"signalbot_go/modules/hugendubel"
	"signalbot_go/modules/news"
	"signalbot_go/modules/periodic"
	"signalbot_go/modules/refectory"
	"signalbot_go/modules/spotify"
	"signalbot_go/modules/tv"
	"signalbot_go/modules/weather"
	"signalbot_go/signalcli"
	signalconsole "signalbot_go/signalcli/drivers/console"
	signaldbus "signalbot_go/signalcli/drivers/dbus"
	signaljsonrpc "signalbot_go/signalcli/drivers/jsonrpc"
	"strings"

	"log/slog"

	"gopkg.in/yaml.v3"
)

// use NewSignalServer to create these structs
// TODO note on concurrency
type SignalServer struct {
	SignalServerCfg
	prefix2module     map[string]string
	acc               *signalcli.Account
	self              string
	modules           map[string]Handler
	log               *slog.Logger
	sockMsgCancel     context.CancelFunc
	sockVirtRcvCancel context.CancelFunc
}

// creates a new signalServer
func NewSignalServer(log *slog.Logger, cfgDir string, dataDir string) (*SignalServer, error) {
	var err error
	// set default
	cfg := SignalServerCfg{
		Dbus: signaldbus.SystemBus,
		UsedDriver: DriverDbus,
	}

	f, err := os.Open(filepath.Join(cfgDir, "main.yaml"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	d := yaml.NewDecoder(f)
	d.KnownFields(true)
	err = d.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	s := SignalServer{
		log:             log,
		modules:         make(map[string]Handler),
		SignalServerCfg: cfg,
	}

	var driver signalcli.Driver
	switch cfg.UsedDriver {
	case DriverDbus:
		driver, err = signaldbus.NewSignalDbusDriver(log.With(), s.Dbus)
		if err != nil {
			return nil, err
		}
	case DriverJsonRpc:
		driver, err = signaljsonrpc.NewSignalJsonRpcDriver(log.With(), cfg.UnixSocket, cfg.SelfNr)
		if err != nil {
			return nil, err
		}
	case DriverConsole:
		driver, err = signalconsole.NewSignalJsonRpcDriver(log.With(), cfg.UnixSocket, cfg.SelfNr)
		if err != nil {
			return nil, err
		}
	}

	s.acc, err = signalcli.NewAccount(log.With(), driver)
	if err != nil {
		return nil, err
	}

	// register functions for handling the messages
	// run the handler in a new goroutine so that new messages can be received
	if err := s.acc.AddMessageHandlerFunc(func(m *signalcli.Message) { go s.handle(m) }); err != nil {
		return nil, err
	}
	if err := s.acc.AddSyncMessageHandlerFunc(func(m *signalcli.SyncMessage) { go s.handle(&m.Message) }); err != nil {
		return nil, err
	}

	// todoMod register modules
	if _, ok := cfg.Handlers["help"]; ok {
		if s.modules["help"], err = NewHelp(log.With("module", "help"), filepath.Join(cfgDir, "help"), s.Handlers, s.self); err != nil {
			return nil, fmt.Errorf("'help' module: %v", err)
		}
	}

	if _, ok := cfg.Handlers["cmd"]; ok {
		if s.modules["cmd"], err = cmd.NewCmd(log.With("module", "cmd"), filepath.Join(cfgDir, "cmd")); err != nil {
			return nil, fmt.Errorf("'cmd' module: %v", err)
		}
	}
	if _, ok := cfg.Handlers["periodic"]; ok {
		if s.modules["periodic"], err = periodic.NewPeriodic(log.With("module", "periodic"), filepath.Join(cfgDir, "periodic")); err != nil {
			return nil, fmt.Errorf("'periodic' module: %v", err)
		}
	}
	if _, ok := cfg.Handlers["refectory"]; ok {
		if s.modules["refectory"], err = refectory.NewRefectory(log.With("module", "refectory"), filepath.Join(cfgDir, "refectory")); err != nil {
			return nil, fmt.Errorf("'refectory' module: %v", err)
		}
	}
	if _, ok := cfg.Handlers["weather"]; ok {
		if s.modules["weather"], err = weather.NewWeather(log.With("module", "weather"), filepath.Join(cfgDir, "weather")); err != nil {
			return nil, fmt.Errorf("'weather' module: %v", err)
		}
	}
	if _, ok := cfg.Handlers["tv"]; ok {
		if s.modules["tv"], err = tv.NewTv(log.With("module", "tv"), filepath.Join(cfgDir, "tv")); err != nil {
			return nil, fmt.Errorf("'tv' module: %v", err)
		}
	}
	if _, ok := cfg.Handlers["buechertreff"]; ok {
		if s.modules["buechertreff"], err = buechertreff.NewBuechertreff(log.With("module", "buechertreff"), filepath.Join(cfgDir, "buechertreff")); err != nil {
			return nil, fmt.Errorf("'buechertreff' module: %v", err)
		}
	}
	if _, ok := cfg.Handlers["freezer"]; ok {
		if s.modules["freezer"], err = freezer.NewFreezer(log.With("module", "freezer"), filepath.Join(cfgDir, "freezer")); err != nil {
			return nil, fmt.Errorf("'freezer' module: %v", err)
		}
	}
	if _, ok := cfg.Handlers["fernsehserien"]; ok {
		if s.modules["fernsehserien"], err = fernsehserien.NewFernsehserien(log.With("module", "fernsehserien"), filepath.Join(cfgDir, "fernsehserien")); err != nil {
			return nil, fmt.Errorf("'fernsehserien' module: %v", err)
		}
	}
	if _, ok := cfg.Handlers["news"]; ok {
		if s.modules["news"], err = news.NewNews(log.With("module", "news"), filepath.Join(cfgDir, "news")); err != nil {
			return nil, fmt.Errorf("'news' module: %v", err)
		}
	}
	if _, ok := cfg.Handlers["hugendubel"]; ok {
		if s.modules["hugendubel"], err = hugendubel.NewHugendubel(log.With("module", "hugendubel"), filepath.Join(cfgDir, "hugendubel")); err != nil {
			return nil, fmt.Errorf("'hugendubel' module: %v", err)
		}
	}
	if _, ok := cfg.Handlers["spotify"]; ok {
		if s.modules["spotify"], err = spotify.NewSpotify(log.With("module", "spotify"), filepath.Join(cfgDir, "spotify")); err != nil {
			return nil, fmt.Errorf("'spotify' module: %v", err)
		}
	}

	// generate prefix2Module
	s.prefix2module = make(map[string]string, len(s.Handlers))
	for name, v := range s.Handlers {
		for _, p := range v.Prefixes {
			s.prefix2module[p] = name
		}
	}

	if err := s.Validate(); err != nil {
		return nil, err
	}

	return &s, nil
}

// check if signalserver is in valid state
func (s *SignalServer) Validate() error {
	if err := s.SignalServerCfg.Validate(); err != nil {
		return err
	}
	// TODO
	// for name := range s.Handlers {
	// 	if _, ok := s.modules[name]; !ok {
	// 		return fmt.Errorf("Trying to register unknown module: %v", name)
	// 	}
	// }
	return nil
}

func (s *SignalServer) startPortSendMsg(ctx context.Context) error {
	listen, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.PortSendMsg))
	if err != nil {
		return err
	}
	go func(listen net.Listener) {
		defer listen.Close()
		running := true
		for running {
			select {
			case <-ctx.Done():
				running = false
			default:
				conn, err := listen.Accept()
				if err != nil {
					s.log.Error("SendMsg: Error accepting on socket", "error", err)
				}
				s.log.Info(fmt.Sprintf("SendMsg: Connected with %s", conn.RemoteAddr().String()))
				go func(conn net.Conn) {
					defer conn.Close()
					m, err := signalcli.NewMessageFromReader(conn, s.self)
					if err != nil || m == nil {
						s.log.Error("SendMsg: Error on reading message from socket", "error", err)
					}
					s.log.Info(fmt.Sprintf("SendMsg: received: %v", m))
					_, err = s.acc.SendGeneric(m.Message, m.Attachments, m.Sender, m.GroupId, false)
					if err != nil {
						s.log.Error("SendMsg: Error on sending message", "error", err)
					}
				}(conn)
			}
		}
	}(listen)
	return nil
}

func (s *SignalServer) startPortVirtRcv(ctx context.Context) error {
	listen, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.PortVirtRcvMsg))
	if err != nil {
		return err
	}
	go func(listen net.Listener) {
		defer listen.Close()
		running := true
		for running {
			select {
			case <-ctx.Done():
				running = false
			default:
				conn, err := listen.Accept()
				if err != nil {
					s.log.Error("VirtRcv: Error accepting on socket", "error", err)
				}
				s.log.Info(fmt.Sprintf("VirtRcv: Connected with %s", conn.RemoteAddr().String()))
				go func(conn net.Conn) {
					defer conn.Close()
					m, err := signalcli.NewMessageFromReader(conn, s.self)
					if err != nil || m == nil {
						s.log.Error("VirtRcv: Error on reading message from socket", "error", err)
					}
					s.log.Info(fmt.Sprintf("VirtRcv: received: %v", m))
					s.handle(m)
				}(conn)
			}
		}
	}(listen)
	return nil
}

// starts the signalserver asynchronously. To fully cleanup call
// signalserver.close()
func (s *SignalServer) Start() error {
	s.acc.ListenForSignals()

	var ctx context.Context

	ctx, s.sockMsgCancel = context.WithCancel(context.Background())
	if err := s.startPortSendMsg(ctx); err != nil {
		s.acc.Close()
		return err
	}

	ctx, s.sockVirtRcvCancel = context.WithCancel(context.Background())
	if err := s.startPortVirtRcv(ctx); err != nil {
		s.sockMsgCancel()
		s.acc.Close()
		return err
	}

	for _, mod := range s.modules {
		if err := mod.Start(s.handle); err != nil {
			return err
		}
	}

	return nil
}

// stops and closes the signalserver. After calling this, the signalserver
// cannot be started again. Please construct a new one with NewSignalServer.
func (s *SignalServer) Close() {
	for _, mod := range s.modules {
		mod.Close(s.handle)
	}

	s.sockVirtRcvCancel()
	s.sockMsgCancel()
	s.acc.Close()
}

// handle a complete signalmessage
func (s *SignalServer) handle(m *signalcli.Message) {
	// unwrap -r
	if m.Message == "-r" {
		if m.GroupId != nil {
			var err error
			m.Message, err = s.acc.GetGroupName(m.GroupId)
			if err != nil {
				s.log.Warn(fmt.Sprintf("could not retreive the groupname of %v. %v", m.GroupId, err))
				return
			}
		} else {
			return // -r cannot be a valid module
		}
	}

	// split the message and handle the different commands

	// split at "\n" as well as "|"
	if m.Message != "" {
		scanner := bufio.NewScanner(strings.NewReader(m.Message))
		scanner.Split(splitLines)
		for scanner.Scan() {
			mLine := *m // copy construct like
			mLine.Message = scanner.Text()
			s.handleLine(&mLine)
		}
	} else {
		s.handleLine(m)
	}
}

// handle the signalmessage as single command
func (s *SignalServer) handleLine(m *signalcli.Message) {
	// TODO alias
	prefix, remainingMsg, _ := strings.Cut(m.Message, " ")
	module, set := s.prefix2module[prefix]
	if !set {
		return
	}

	// check authorization
	{
		handler, set := s.Handlers[module]
		if !set {
			s.log.Warn(fmt.Sprintf("No handler found for module %v", module))
			return
		}
		if err := handler.Access.Check(m.Sender, m.Chat); err != nil {
			s.log.Info("Accesscontrol blocked.", "Error", err)
			return
		}
	}
	// at this point the user is authorized for this module

	s.log.Info(fmt.Sprintf("Handling: %v -> %v", m, remainingMsg))
	m.Message = remainingMsg
	if mod,ok := s.modules[module]; !ok {
		s.log.Error("Trying to call module which is registered but not available", "module", module)
	} else {
		mod.Handle(m, s.acc, s.handle)
	}
}
