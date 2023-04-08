package signalserver

import (
	"fmt"
	"signalbot_go/internal/signalsender"
	"signalbot_go/signaldbus"
	"strings"
)

// can handle a signal-message
type Handler interface {
	Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message))
	Start(virtRcv func(*signaldbus.Message)) error
	Close(virtRcv func(*signaldbus.Message))
}

// config for a handler. Can be parsed from yaml
// TODO note on concurrency
type HandlerCfg struct {
	Prefixes []string              `yaml:"prefixes"`
	Help     string                `yaml:"help"`
	Access   Accesscontrol         `yaml:"access"`
}

// validate the stored data
func (c *HandlerCfg) Validate() error {
	for _, p := range c.Prefixes {
		if strings.ContainsRune(p, ' ') {
			return fmt.Errorf("Invalid prefix: %v (cannot contain space)", p)
		}
	}
	return c.Access.Validate()
}
