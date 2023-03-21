package signalsender

import "signalbot_go/signaldbus"

// an interface which allows to send data to signal
type SignalSender interface {
	// send message to arbitrary recipient.If groupID is empty, send to
	// recipient. If groupID is set, the message is sent to the group (and the
	// recipient is ignored)
	SendGeneric(message string, attachments []string, recipient string, groupID []byte) (timestamp int64, err error)
	// respond to a certain message. The recipient/groupID will be extracted from the message
	Respond(message string, attachments []string, m *signaldbus.Message) (timestamp int64, err error)
}
