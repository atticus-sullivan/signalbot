package signalserver

// todo: dirctly call handler for periodic stuff s.handle(m)
// todo open websocket for virtually sending messages (calls s.handle(m)). Message is passed via json (todo annotations to message struct for (un)marshalling)

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"signalbot_go/modules/cmd"
	"signalbot_go/signaldbus"
	"strings"

	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

// use NewSignalServer to create these structs
// TODO note on concurrency
type SignalServer struct {
	SignalServerCfg
	prefix2module map[string]string
	acc           *signaldbus.Account
	modules       map[string]Handler
	log           *slog.Logger
}

// creates a new signalServer
func NewSignalServer(log *slog.Logger, cfgDir string, dataDir string) (*SignalServer, error) {
	var err error
	cfg := SignalServerCfg{Dbus: signaldbus.SystemBus}

	f, err := os.Open(filepath.Join(cfgDir, "main.yaml"))
	if err != nil {
		return nil, err
	}
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
	s.acc.AddMessageHandlerFunc(func(m *signaldbus.Message) { go s.handle(m) })
	s.acc.AddSyncMessageHandlerFunc(func(m *signaldbus.SyncMessage) { go s.handle(&m.Message) })

	// todoMod register modules
	if s.modules["cmd"], err = cmd.NewCmd(log.With(), filepath.Join(cfgDir, "cmd")); err != nil {
		return nil, err
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
	for name := range(s.Handlers) {
		if _,ok := s.modules[name]; !ok {
			return fmt.Errorf("Trying to register unknown module: %v", name)
		}
	}
	return nil
}

// starts the signalserver asynchronously. To fully cleanup call
// signalserver.close()
func (s *SignalServer) Start() {
	s.acc.ListenForSignals()
}

// stops and closes the signalserver. After calling this, the signalserver
// cannot be started again. Please construct a new one with NewSignalServer.
func (s *SignalServer) Close() {
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
	module := s.prefix2module[prefix]

	// check authorization
	{
		user_allow, set := s.Handlers[module].Access.AccessSet[m.Sender]
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

	m.Message = remainingMsg
	s.modules[module].Handle(m, s.acc)
}
