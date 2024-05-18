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

func validChat(chat string) bool {
	return validHexstring(chat) || validPhoneNr(chat)
}
