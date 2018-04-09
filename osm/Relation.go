package osm

import (
	"time"
)

type RelationMember struct {
	Type      string `xml:"type,attr"`
	Reference int64  `xml:"ref,attr"`
	Role      string `xml:"role,attr"`
}

type Relation struct {
	Id        int64            `xml:"id,attr"`
	Version   int              `xml:"version,attr,omitempty"`
	Timestamp *time.Time       `xml:"timestamp,attr,omitempty"`
	Changeset int64            `xml:"changeset,attr,omitempty"`
	UserId    int64            `xml:"uid,attr,omitempty"`
	UserName  string           `xml:"user,attr,omitempty"`
	Members   []RelationMember `xml:"member"`
	Tags      []Tag            `xml:"tag"`
}

func (r *Relation) DropVersion() {
	r.Version = 0
}

func (r *Relation) DropTimestamp() {
	r.Timestamp = nil
}

func (r *Relation) DropChangeset() {
	r.Changeset = int64(0)
}
