package differ

import (
	"strings"
)

type diffStringer interface {
	comparable
	AddString() string
	RemString() string
}

type Differ[S comparable, T comparable, U diffStringer] map[S]map[T][]U

func (d *Differ[S,T,U]) Diff(l1 S, l2 T, dataB []U) (string) {
	first := true
	a, ok := (*d)[l1]
	if !ok {
		// everything is new
		builder := strings.Builder{}
		for _, dB := range dataB {
			if !first {
				builder.WriteRune('\n')
			} else {
				first = false
			}
			builder.WriteString(dB.AddString())
		}
		return builder.String()
	}
	dataA, ok := a[l2]
	if !ok {
		// everything is new
		builder := strings.Builder{}
		for _, dB := range dataB {
			if !first {
				builder.WriteRune('\n')
			} else {
				first = false
			}
			builder.WriteString(dB.AddString())
		}
		return builder.String()
	}
	builder := strings.Builder{}


	for _, dA := range(dataA) {
		found := false
		for _, dB := range(dataB) {
			if dA == dB {
				found = true
				break
			}
		}
		if !found {
			// is in A but not in B
			str := dA.RemString()
			if str != "" {
				if !first {
					builder.WriteRune('\n')
				} else {
					first = false
				}
				builder.WriteString(str)
			}
		}
	}
	for _, dB := range(dataB) {
		found := false
		for _, dA := range(dataA) {
			if dA == dB {
				found = true
				break
			}
		}
		if !found {
			// is in B but not in A
			str := dB.AddString()
			if str != "" {
				builder.WriteString(str)
				if !first {
					builder.WriteRune('\n')
				} else {
					first = false
				}
			}
		}
	}

	return builder.String()
}

func (d *Differ[S,T,U]) DiffStore(l1 S, l2 T, dataB []U) (string) {
	if d == nil {
		return ""
	}

	// diff
	resp := d.Diff(l1, l2, dataB)

	// store
	if _,ok := (*d)[l1]; !ok {
		(*d)[l1] = make(map[T][]U)
	}
	(*d)[l1][l2] = dataB

	return resp
}
