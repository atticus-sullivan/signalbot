package signalserver

import "fmt"

// stores if the access is granted or blocked
type Capability string

const (
	Allow Capability = "Allow"
	Block Capability = "Block"
)

// Capability is a stringer
func (c Capability) String() string {
	return string(c)
}

// Accesscontol struct with Capabilities of a user
// Can be parsed from a yaml file
// TODO note on concurrency
type AccesscontrolUser struct {
	Def       Capability      `yaml:"default"`
	AccessSet map[string]bool `yaml:"accessSet"`
}

// Validate if stored information is valid
func (a *AccesscontrolUser) Validate() error {
	if a.Def != Allow && a.Def != Block {
		return fmt.Errorf("Invalid default value: %v", a.Def)
	}
	for user := range a.AccessSet {
		if !validPhoneNr(user) {
			return fmt.Errorf("Invalid user: %v (must be a phoneNuber)", user)
		}
	}
	return nil
}

// Accesscontol struct with Capabilities of a user in a specific group/chat.
// Can be parsed from a yaml file
// TODO note on concurrency
type AccesscontrolUserChat struct {
	Def       Capability                 `yaml:"default"`
	AccessSet map[string]map[string]bool `yaml:"accessSet"`
}

// Validate if stored information is valid
func (a *AccesscontrolUserChat) Validate() error {
	if a.Def != Allow && a.Def != Block {
		return fmt.Errorf("Invalid default value: %v", a.Def)
	}
	for user, m := range a.AccessSet {
		if !validPhoneNr(user) {
			return fmt.Errorf("Invalid user: %v (must be a phoneNuber)", user)
		}
		for chat := range m {
			if !validChat(chat) {
				return fmt.Errorf("Invalid groupID/chat: %v (must be a hex-string or 'direct' for PNs)", chat)
			}
		}
	}
	return nil
}
