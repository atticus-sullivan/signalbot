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
	a, ok := (*d)[l1]
	if !ok {
		// everything is new
		builder := strings.Builder{}
		for _, dB := range dataB {
			builder.WriteString(dB.AddString())
			builder.WriteRune('\n')
		}
		return builder.String()
	}
	dataA, ok := a[l2]
	if !ok {
		// everything is new
		builder := strings.Builder{}
		for _, dB := range dataB {
			builder.WriteString(dB.AddString())
			builder.WriteRune('\n')
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
			builder.WriteString(dA.RemString())
			builder.WriteRune('\n')
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
			builder.WriteString(dB.AddString())
			builder.WriteRune('\n')
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
