package osm

import (
	"sort"
)

func SearchInt64s(a []int64, x int64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Int64Slice) Sort() { sort.Sort(p) }

func (p Int64Slice) Search(x int64) int {
	return SearchInt64s(p, x)
}

func (p Int64Slice) Contains(x int64) bool {
	i := p.Search(x)
	return i < p.Len() && p[i] == x
}

func NewInt64Slice(length int, capacity int) Int64Slice {
	return Int64Slice(make([]int64, length, capacity))
}
