package osm

type Int64Set map[int64]struct{}

func (set Int64Set) Add(x int64) {
	set[x] = struct{}{}
}

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

func NewInt64Set() Int64Set {
	return make(map[int64]struct{})
}
