package sdiffer

import (
	"reflect"
)

// Sorter sort slice before comparison to do disordered comparison.
type Sorter interface {

	// Match checks if a field should use this comparator.
	Match(fieldPath string) bool

	// Less calculate if 'a' is less than 'b'.
	//
	// For example:
	// type Integer struct {
	// 		val int
	// }
	// less := func(a, b interface{}) bool {
	// 		i, j := a.(Integer), b.(Integer)
	// 		return i.val < j.val
	// }
	Less(a, b interface{}) bool
}

func qsort(slice reflect.Value, less func(a, b interface{}) bool) {
	doQsort(slice, less, 0, slice.Len()-1)
}

func doQsort(slice reflect.Value, less func(a, b interface{}) bool, start, end int) {
	if start < end {
		m := slice.Index(start).Interface()
		l, r := start, end
		for l < r {
			for l < r && less(m, slice.Index(r).Interface()) {
				r--
			}
			if l < r {
				slice.Index(l).Set(slice.Index(r))
				l++
			}
			for l < r && less(slice.Index(l).Interface(), m) {
				l++
			}
			if l < r {
				slice.Index(r).Set(slice.Index(l))
				r--
			}
		}
		slice.Index(l).Set(reflect.ValueOf(m))
		doQsort(slice, less, start, l-1)
		doQsort(slice, less, l+1, end)
	}
}
