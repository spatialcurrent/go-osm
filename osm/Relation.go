package osm

import (
//"time"
)

type RelationMember struct {
	Type      string `xml:"type,attr"`
	Reference int64  `xml:"ref,attr"`
	Role      string `xml:"role,attr"`
}

type Relation struct {
	TaggedElement
	Members []RelationMember `xml:"member"`
	Tags    []Tag            `xml:"tag"`
}

func (r *Relation) DropAttributes(output Output) {
	if output.DropVersion {
		r.DropVersion()
	}

	if output.DropTimestamp {
		r.DropTimestamp()
	}

	if output.DropChangeset {
		r.DropChangeset()
	}

	if output.DropUserId {
		r.DropUserId()
	}

	if output.DropUserName {
		r.DropUserName()
	}
}
