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

func (r Relation) HasKey(key string) bool {
	for _, t := range r.Tags {
		if key == t.Key {
			return true
		}
	}
	return false
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

func (r *Relation) DropVersion() {
	r.Version = 0
}

func (r *Relation) DropTimestamp() {
	r.Timestamp = nil
}

func (r *Relation) DropChangeset() {
	r.Changeset = int64(0)
}

func (r *Relation) DropUserId() {
	r.UserId = int64(0)
}

func (r *Relation) DropUserName() {
	r.UserName = ""
}

func (r *Relation) TagsAsMap() map[string]interface{} {
	m := map[string]interface{}{}
	for _, t := range r.Tags {
		m[t.Key] = t.Value
	}
	return m
}
