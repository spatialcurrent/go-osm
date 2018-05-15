package osm

import (
	"sort"
)

// StringSet is a logical set of string values using a map[string]struct{} backend.
// The use of a string -> empty struct backend provides a higher write performance versus a slice backend.
type StringSet map[string]struct{}

func (set StringSet) Contains(x string) bool {
	_, ok := set[x]
	return ok
}

func (set StringSet) Len() int {
	return len(set)
}

// Add is a variadic function to add values to a set
//	- https://gobyexample.com/variadic-functions
func (set StringSet) Add(values ...string) {
	for _, v := range values {
		set[v] = struct{}{}
	}
}

func (set StringSet) Union(values interface{}) StringSet {
	union := NewStringSet()
	for x := range set {
		union.Add(x)
	}
	switch values.(type) {
	case []string:
		for _, v := range values.([]string) {
			union.Add(v)
		}
	case StringSet:
		for v := range values.(StringSet) {
			union.Add(v)
		}
	}
	return union
}

func (set StringSet) Intersect(values interface{}) StringSet {
	intersection := NewStringSet()
	switch values.(type) {
	case []string:
		for _, v := range values.([]string) {
			if set.Contains(v) {
				intersection.Add(v)
			}
		}
	case StringSet:
		for v := range values.(StringSet) {
			if set.Contains(v) {
				intersection.Add(v)
			}
		}
	}

	return intersection
}

// Slice returns a slice representation of this set.
// If parameter sorted is true, then sorts the values using natural sort order.
func (set StringSet) Slice(sorted bool) sort.StringSlice {
	slice := sort.StringSlice(make([]string, len(set)))
	for x := range set {
		slice = append(slice, x)
	}
	if sorted {
		slice.Sort()
	}
	return slice
}

// NewStringSet returns a new StringSet.
func NewStringSet() StringSet {
	return make(map[string]struct{})
}
