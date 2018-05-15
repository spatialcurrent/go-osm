package osm

import (
	"sort"
)

// SearchUInt64s searches slice "a" for the value "x" and returns the index.
// The index may be a false positive, so it is important to check the target value at the returned index.
func SearchUInt64s(a []uint64, x uint64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// UInt64Slice is a type alias for a slice of int64.
type UInt64Slice []uint64

func (p UInt64Slice) Len() int           { return len(p) }
func (p UInt64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p UInt64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p UInt64Slice) Sort() { sort.Sort(p) }

func (p UInt64Slice) Search(x uint64) int {
	return SearchUInt64s(p, x)
}

func (p UInt64Slice) Contains(x uint64) bool {
	i := p.Search(x)
	return i < p.Len() && p[i] == x
}

// NewUInt64Slice returns a new UInt64Slice with length and capacity given as parameters.
func NewUInt64Slice(length int, capacity int) UInt64Slice {
	return UInt64Slice(make([]uint64, length, capacity))
}
