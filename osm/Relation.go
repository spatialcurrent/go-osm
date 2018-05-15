package osm

import (
//"time"
)

type RelationMember struct {
	Type      string `xml:"type,attr"`
	Reference uint64 `xml:"ref,attr"`
	Role      string `xml:"role,attr"`
}

type Relation struct {
	TaggedElement
	Members []RelationMember `xml:"member"`
}

func (r Relation) NumberOfMembers() int {
	return len(r.Members)
}

func NewRelation() *Relation {
	return &Relation{
		TaggedElement: TaggedElement{Element: Element{}},
		Members:       []RelationMember{},
	}
}
