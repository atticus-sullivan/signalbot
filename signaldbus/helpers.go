package signaldbus

// send message to arbitrary recipient.If groupID is empty, send to
// recipient. If groupID is set, the message is sent to the group (and the
// recipient is ignored)
func (s *Account) SendGeneric(message string, attachments []string, recipient string, groupID []byte, notify bool) (timestamp int64, err error) {
	if len(groupID) > 0 {
		// send group message ignoring recipient
		return s.SendGroupMessage(message, attachments, groupID)
	} else {
		// send normal personal message
		return s.SendMessage(message, attachments, recipient, notify)
	}
}

// respond to a certain message. The recipient/groupID will be extracted from the message
func (s *Account) Respond(message string, attachments []string, m *Message, notify bool) (timestamp int64, err error) {
	dst := m.Sender
	if m.Sender == s.selfNr {
		dst = m.Receiver
	}
	return s.SendGeneric(message, attachments, dst, m.GroupId, notify)
}
