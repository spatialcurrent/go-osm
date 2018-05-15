package osm

// UInt64Set is a logical set of int64 values using a map[uint64]struct{} backend.
// The use of a uint64 -> empty struct backend provides a higher write performance versus a slice backend.
type UInt64Set map[uint64]struct{}

// Add adds parameter x to the set.
func (set UInt64Set) Add(x uint64) {
	set[x] = struct{}{}
}

// Slice returns a slice representation of this set.
// If parameter sorted is true, then sorts the values using natural sort order.
func (set UInt64Set) Slice(sorted bool) UInt64Slice {
	slice := NewUInt64Slice(0, len(set))
	for x := range set {
		slice = append(slice, x)
	}
	if sorted {
		slice.Sort()
	}
	return slice
}

// NewUInt64Set returns a new Int64Set.
func NewUInt64Set() UInt64Set {
	return make(map[uint64]struct{})
}
