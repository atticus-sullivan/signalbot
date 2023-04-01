package signalserver

// todo: dirctly call handler for periodic stuff s.handle(m)
// todo open websocket for virtually sending messages (calls s.handle(m)). Message is passed via json (todo annotations to message struct for (un)marshalling)

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"signalbot_go/modules/cmd"
	"signalbot_go/modules/periodic"
	"signalbot_go/modules/refectory"
	"signalbot_go/modules/tv"
	"signalbot_go/modules/weather"
	"signalbot_go/signaldbus"
	"strings"

	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

// use NewSignalServer to create these structs
// TODO note on concurrency
type SignalServer struct {
	SignalServerCfg
	prefix2module     map[string]string
	acc               *signaldbus.Account
	modules           map[string]Handler
	log               *slog.Logger
	sockMsgCancel     context.CancelFunc
	sockVirtRcvCancel context.CancelFunc
}

// creates a new signalServer
func NewSignalServer(log *slog.Logger, cfgDir string, dataDir string) (*SignalServer, error) {
	var err error
	cfg := SignalServerCfg{Dbus: signaldbus.SystemBus}

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

	s := SignalServer{
		log:             log,
		modules:         make(map[string]Handler),
		SignalServerCfg: cfg,
	}

	s.acc, err = signaldbus.NewAccount(log.With(), s.Dbus)
	if err != nil {
		return nil, err
	}
	// register functions for handling the messages
	// run the handler in a new goroutine so that new messages can be received
	if err := s.acc.AddMessageHandlerFunc(func(m *signaldbus.Message) { go s.handle(m) }); err != nil {
		return nil, err
	}
	if err := s.acc.AddSyncMessageHandlerFunc(func(m *signaldbus.SyncMessage) { go s.handle(&m.Message) }); err != nil {
		return nil, err
	}

	// todoMod register modules
	if _,ok := cfg.Handlers["cmd"]; ok {
		if s.modules["cmd"], err = cmd.NewCmd(log.With(), filepath.Join(cfgDir, "cmd")); err != nil {
			return nil, fmt.Errorf("'cmd' module: %v", err)
		}
	}
	if _,ok := cfg.Handlers["periodic"]; ok {
		if s.modules["periodic"], err = periodic.NewPeriodic(log.With(), filepath.Join(cfgDir, "periodic")); err != nil {
			return nil, fmt.Errorf("'periodic' module: %v", err)
		}
	}
	if _,ok := cfg.Handlers["refectory"]; ok {
		if s.modules["refectory"], err = refectory.NewRefectory(log.With(), filepath.Join(cfgDir, "refectory")); err != nil {
			return nil, fmt.Errorf("'refectory' module: %v", err)
		}
	}
	if _,ok := cfg.Handlers["weather"]; ok {
		if s.modules["weather"], err = weather.NewWeather(log.With(), filepath.Join(cfgDir, "weather")); err != nil {
			return nil, fmt.Errorf("'weather' module: %v", err)
		}
	}
	if _,ok := cfg.Handlers["tv"]; ok {
		if s.modules["tv"], err = tv.NewTv(log.With(), filepath.Join(cfgDir, "tv")); err != nil {
			return nil, fmt.Errorf("'tv' module: %v", err)
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
	for name := range s.Handlers {
		if _, ok := s.modules[name]; !ok {
			return fmt.Errorf("Trying to register unknown module: %v", name)
		}
	}
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
					m, err := signaldbus.NewMessageFromReader(conn)
					if err != nil || m == nil {
						s.log.Error("SendMsg: Error on reading message from socket", "error", err)
					}
					s.log.Info(fmt.Sprintf("SendMsg: received: %v", m))
					_, err = s.acc.SendGeneric(m.Message, m.Attachments, m.Sender, m.GroupId)
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
					m, err := signaldbus.NewMessageFromReader(conn)
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
func (s *SignalServer) handle(m *signaldbus.Message) {
	if allow, set := s.Access.AccessSet[m.Sender]; (!set && s.Access.Def == Block) || (set && !allow) {
		return
	}
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
	scanner := bufio.NewScanner(strings.NewReader(m.Message))
	scanner.Split(splitLines)
	for scanner.Scan() {
		mLine := *m // copy construct like
		mLine.Message = scanner.Text()
		s.handleLine(m)
	}
}

// handle the signalmessage as single command
func (s *SignalServer) handleLine(m *signaldbus.Message) {
	prefix, remainingMsg, _ := strings.Cut(m.Message, " ")
	module, set := s.prefix2module[prefix]
	if !set {
		s.log.Warn(fmt.Sprintf("Unknown prefix %v", prefix))
		return
	}

	// check authorization
	{
		handler, set := s.Handlers[module]
		if !set {
			s.log.Warn(fmt.Sprintf("No handler found for module %v", module))
			return
		}
		user_allow, set := handler.Access.AccessSet[m.Sender]
		if !set && s.Handlers[module].Access.Def == Block {
			s.log.Info(fmt.Sprintf("Accesscontrol blocked. Set: %v, Default: %v", set, s.Handlers[module].Access.Def))
			return
		}
		allow, set := user_allow[hex.EncodeToString(m.GroupId)]
		if (!set && s.Handlers[module].Access.Def == Block) || (set && !allow) {
			s.log.Info(fmt.Sprintf("Accesscontrol blocked. Set: %v, allow: %v, user: %v, module: %v", set, allow, m.Sender, hex.EncodeToString(m.GroupId)))
			return
		}
	}
	// at this point the user is authorized for this module

	s.log.Info(fmt.Sprintf("Handling: %v -> %v", m, remainingMsg))
	m.Message = remainingMsg
	s.modules[module].Handle(m, s.acc, s.handle)
}
