package signalserver

import (
	"bytes"
	"regexp"
)

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func splitLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if j, i := bytes.IndexByte(data, '|'), bytes.IndexByte(data, '\n'); i >= 0 || j >= 0 {
		// We have a full newline-terminated line.
		var idx int
		if j >= 0 && (j < i || i < 0) {
			idx = j
		} else {
			idx = i
		}
		return idx + 1, dropCR(data[0:idx]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

var phoneNrRe *regexp.Regexp = regexp.MustCompile(`^\+[0-9]{3,}$`)

func validPhoneNr(phoneNr string) bool {
	return phoneNrRe.MatchString(phoneNr)
}

var hexStringRe *regexp.Regexp = regexp.MustCompile(`^[[:xdigit:]]+$`)

func validHexstring(hex string) bool {
	return hexStringRe.MatchString(hex)
}

// todo make this chat and store chat instead of string?
type ChatType string

const Direct ChatType = "direct" // todo put thie somewhere else?
func validChat(chat string) bool {
	return validHexstring(chat) || ChatType(chat) == Direct
}
