package osm

type TaggedElement struct {
	Element
	Tags      []Tag `xml:"tag"`
	TagsIndex []int `xml:"-"`
}

func (te TaggedElement) GetTagsIndex() []int {
	return te.TagsIndex
}

func (te *TaggedElement) SetTagsIndex(tagsIndex []int) {
	te.TagsIndex = tagsIndex
}

func (te *TaggedElement) SetTags(tags []Tag) {
	te.Tags = tags
}
