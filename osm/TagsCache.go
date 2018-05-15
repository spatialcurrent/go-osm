package osm

// TagsCache is a cache of tags with a map for reverse lookup
type TagsCache struct {
	Values []Tag                        // slice of cached tags
	Index  map[string]map[string]uint32 // map[key][value] ==> index in Values
}

// AddTag adds a tag to the cache and returns its index
func (tc *TagsCache) AddTag(t Tag) uint32 {
	if _, ok := tc.Index[t.Key]; ok {
		if i, ok := tc.Index[t.Key][t.Value]; ok {
			return i
		} else {
			tc.Values = append(tc.Values, t)
			tc.Index[t.Key][t.Value] = uint32(len(tc.Values) - 1)
		}
	} else {
		tc.Values = append(tc.Values, t)
		tc.Index[t.Key] = map[string]uint32{}
		tc.Index[t.Key][t.Value] = uint32(len(tc.Values) - 1)
	}
	return uint32(len(tc.Values) - 1)
}

// AddTags add the slice of tags to the cache and returns a slice of indicies
func (tc *TagsCache) AddTags(tags []Tag) []uint32 {
	tagIndicies := make([]uint32, len(tags))
	for i, t := range tags {
		tagIndicies[i] = tc.AddTag(t)
	}
	return tagIndicies
}

func (tc *TagsCache) Slice(tagIndicies []uint32) []Tag {
	s := make([]Tag, len(tagIndicies))
	for i, x := range tagIndicies {
		s[i] = Tag{Key: tc.Values[x].Key, Value: tc.Values[x].Value}
	}
	return s
}

func (tc *TagsCache) Map(tagIndicies []uint32) map[string]interface{} {
	m := make(map[string]interface{}, len(tagIndicies))
	for _, x := range tagIndicies {
		m[tc.Values[x].Key] = tc.Values[x].Value
	}
	return m
}

func NewTagsCache() *TagsCache {
	return &TagsCache{
		Values: make([]Tag, 0),
		Index:  map[string]map[string]uint32{},
	}
}
