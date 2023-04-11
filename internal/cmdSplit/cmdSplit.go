package cmdsplit

import (
	"fmt"
	"strings"
)

type state int

const (
	normal state = iota
	normalEscaped
	quoted
	quotedEscaped
)

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
		return nil, fmt.Errorf("Malformed string, open escape sequence or quote in the end")
	}
	ret = append(ret, collectString.String())

	if len(ret) > 0 && ret[len(ret)-1] == "" {
		ret = ret[:len(ret)-1]
	}

	return ret, nil
}

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
		return "", fmt.Errorf("Malformed string, open escape sequence in the end")
	}

	return ret.String(), nil
}
