package osm

type TagsCache struct {
	Values []Tag
	Index  map[string]map[string]int
}

func (tc *TagsCache) AddTag(t Tag) int {
	if _, ok := tc.Index[t.Key]; ok {
		if i, ok := tc.Index[t.Key][t.Value]; ok {
			return i
		} else {
			tc.Values = append(tc.Values, t)
			tc.Index[t.Key][t.Value] = len(tc.Values) - 1
		}
	} else {
		tc.Values = append(tc.Values, t)
		tc.Index[t.Key] = map[string]int{}
		tc.Index[t.Key][t.Value] = len(tc.Values) - 1
	}
	return len(tc.Values) - 1
}

func (tc *TagsCache) AddTags(tags []Tag) []int {
	tagIndicies := make([]int, len(tags))
	for i, t := range tags {
		tagIndicies[i] = tc.AddTag(t)
	}
	return tagIndicies
}

func (tc *TagsCache) Slice(tagIndicies []int) []Tag {
	s := make([]Tag, len(tagIndicies))
	for i, x := range tagIndicies {
		s[i] = Tag{Key: tc.Values[x].Key, Value: tc.Values[x].Value}
	}
	return s
}

func (tc *TagsCache) Map(tagIndicies []int) map[string]interface{} {
	m := map[string]interface{}{}
	for _, x := range tagIndicies {
		m[tc.Values[x].Key] = tc.Values[x].Value
	}
	//for _, t := range te.Tags {
	//	m[t.Key] = t.Value
	//}
	return m
}

func NewTagsCache() *TagsCache {
	return &TagsCache{
		Values: make([]Tag, 0),
		Index:  map[string]map[string]int{},
	}
}
