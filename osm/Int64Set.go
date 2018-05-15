package osm

// Int64Set is a logical set of int64 values using a map[int64]struct{} backend.
// The use of a int64 -> empty struct backend provides a higher write performance versus a slice backend.
type Int64Set map[int64]struct{}

// Add adds parameter x to the set.
func (set Int64Set) Add(x int64) {
	set[x] = struct{}{}
}

// Slice returns a slice representation of this set.
// If parameter sorted is true, then sorts the values using natural sort order.
func (set Int64Set) Slice(sorted bool) Int64Slice {
	slice := NewInt64Slice(0, len(set))
	for x := range set {
		slice = append(slice, x)
	}
	if sorted {
		slice.Sort()
	}
	return slice
}

// NewInt64Set returns a new Int64Set.
func NewInt64Set() Int64Set {
	return make(map[int64]struct{})
}
