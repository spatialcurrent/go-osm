package osm

// TaggedElement is an abstract struct that attaches tags to OSM elements and associated functions
type TaggedElement struct {
	Element            // the extended OSM element
	Tags      []Tag    `xml:"tag"`                                                          // a slice of tags for this element
	TagsIndex []uint32 `xml:"-" parquet:"name=tags, type=U_INT32, repetitiontype=REPEATED"` // a slice of indexes in the Planet TagCache.
}

// GetTagsIndex returns a slice of indexes in the Planet TagCache for this element
func (te TaggedElement) GetTagsIndex() []uint32 {
	return te.TagsIndex
}

// SetTagsIndex sets the indexes in the Planet TagCache to retrieve the associated Tag
func (te *TaggedElement) SetTagsIndex(tagsIndex []uint32) {
	te.TagsIndex = tagsIndex
}

// SetTags sets the elements Tags
func (te *TaggedElement) SetTags(tags []Tag) {
	te.Tags = tags
}
