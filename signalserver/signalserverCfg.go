package signalserver

import (
	"fmt"
	"signalbot_go/signaldbus"
)

// configuration of a signalServer. Can be parsed by yaml
// TODO note on concurrency
type SignalServerCfg struct {
	Dbus     signaldbus.DbusType   `yaml:"dbus"`
	Access   AccesscontrolUser     `yaml:"access"`
	Handlers map[string]HandlerCfg `yaml:"handlers"` // maps name to prefix

	// just to have a place where to define anchors to alias to laster
	Chats []string `yaml:"chats"`
	Users []string `yaml:"users"`
}

// check if stored values are valid
func (c *SignalServerCfg) Validate() error {
	if c.Dbus != signaldbus.SystemBus && c.Dbus != signaldbus.SessionBus {
		return fmt.Errorf("Invalid dbus type")
	}
	if err := c.Access.Validate(); err != nil {
		return err
	}
	for _, h := range c.Handlers {
		if err := h.Validate(); err != nil {
			return err
		}
	}
	// no validation of Chats and Users as it is only for anchors in the config
	return nil
}
