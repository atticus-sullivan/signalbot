package cmdsplit

import (
	"errors"
	"strings"
)

// errors
var (
	ErrTrailingQuoteEscape error = errors.New("Malformed string, open escape sequence or quote in the end")
	ErrTrailingEscape      error = errors.New("Malformed string, open escape sequence in the end")
)

// possible states of the DFA
type state int

const (
	normal state = iota
	normalEscaped
	quoted
	quotedEscaped
)

// split the string into a slice if strings. Splitting is done similar to the
// argument splitting in the shell (in general at spaces). ' and " are read as
// quotation marks and \ as escape character.
func Split(s string) ([]string, error) {
	ret := make([]string, 0)
	collectString := strings.Builder{}
	st := normal

	q := rune(0)
	for _, c := range s {

		switch st {

		case normal:
			switch c {
			case '\\':
				st = normalEscaped
			case '\'':
				q = '\''
				st = quoted
			case '"':
				q = '"'
				st = quoted
			case ' ':
				ret = append(ret, collectString.String())
				collectString.Reset()
			default:
				collectString.WriteRune(c)
			}

		case normalEscaped:
			collectString.WriteRune(c)
			st = normal

		case quotedEscaped:
			collectString.WriteRune(c)
			st = quoted

		case quoted:
			switch {
			case c == '\\':
				st = quotedEscaped
			case c == q:
				st = normal
			default:
				collectString.WriteRune(c)
			}
		}
	}

	if st != normal {
		return nil, ErrTrailingQuoteEscape
	}
	ret = append(ret, collectString.String())

	if len(ret) > 0 && ret[len(ret)-1] == "" {
		ret = ret[:len(ret)-1]
	}

	return ret, nil
}

// removes one layer of escaping from the string. This means that something
// like
// hello\ world
// becomes
// hello world
func Unescape(s string) (string, error) {
	ret := strings.Builder{}
	st := normal

	for _, c := range s {

		switch st {

		case normal:
			switch c {
			case '\\':
				st = normalEscaped
			default:
				ret.WriteRune(c)
			}

		case normalEscaped:
			ret.WriteRune(c)
			st = normal
		}
	}

	if st != normal {
		return "", ErrTrailingEscape
	}

	return ret.String(), nil
}
