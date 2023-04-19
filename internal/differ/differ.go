package differ

import (
	"strings"
)

// interface comparable to the stringer but has Add-/RemString functions to
// format the text shown when the element was added/removed.
// In order to be able to compare it, it needs to implement comparable!
type diffStringer interface {
	comparable
	AddString() string
	RemString() string
}

type equaler[T any] interface {
	Equals(other T) bool
}

type diffStringerEqualer[T any] interface {
	equaler[T]
	diffStringer
}

// stores the last state of the list-collection of U. Finding the
// list-collection is done by following a path of two parameters S and T in a
// kinda tree.
// This object can then be used to diff an arbitrary state with the stored one
// and/or store the given state afterwards.
type Differ[S comparable, T comparable, U diffStringerEqualer[U]] map[S]map[T][]U

// Generate a diff between the stored state (found by following `l1` and `l2`)
// the and provided state `dataB`.
// The output is generated with the help of the AddString/RemString functions
// of the U elements.
func (d *Differ[S, T, U]) Diff(l1 S, l2 T, dataB []U) string {
	first := true // used to omit the leading newline in the first iteration
	a, ok := (*d)[l1]
	if !ok {
		// everything is new as the path wasn't found
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
		// everything is new as the path wasn't found
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

	// check if elements of dataA are contained in dataB
	for _, dA := range dataA {
		found := false
		for _, dB := range dataB {
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

	// reverse: check if elements of dataB are contained in dataA
	for _, dB := range dataB {
		found := false
		for _, dA := range dataA {
			if dA == dB {
				found = true
				break
			}
		}
		if !found {
			// is in B but not in A
			str := dB.AddString()
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

	return builder.String()
}

// same as `Diff` but automatically stores the provided `dataB` in the `Differ`
// after comparing.
func (d *Differ[S, T, U]) DiffStore(l1 S, l2 T, dataB []U) string {
	if d == nil {
		return ""
	}

	// diff
	resp := d.Diff(l1, l2, dataB)

	// store
	if _, ok := (*d)[l1]; !ok {
		(*d)[l1] = make(map[T][]U)
	}
	(*d)[l1][l2] = dataB

	return resp
}
