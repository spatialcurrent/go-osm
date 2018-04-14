package osm

import (
	"time"
)

type NodeReference struct {
	Reference int64 `xml:"ref,attr"`
}

type Way struct {
	Id             int64           `xml:"id,attr"`
	Version        int             `xml:"version,attr,omitempty"`
	Timestamp      *time.Time      `xml:"timestamp,attr,omitempty"`
	Changeset      int64           `xml:"changeset,attr,omitempty"`
	UserId         int64           `xml:"uid,attr,omitempty"`
	UserName       string          `xml:"user,attr,omitempty"`
	NodeReferences []NodeReference `xml:"nd"`
	Tags           []Tag           `xml:"tag"`
}

func (w Way) NumberOfNodes() int {
	return len(w.NodeReferences)
}

func (w Way) HasKey(key string) bool {
	for _, t := range w.Tags {
		if key == t.Key {
			return true
		}
	}
	return false
}

func (w *Way) DropVersion() {
	w.Version = 0
}

func (w *Way) DropTimestamp() {
	w.Timestamp = nil
}

func (w *Way) DropChangeset() {
	w.Changeset = int64(0)
}

func (w *Way) DropUid() {
	w.UserId = int64(0)
}

func (w *Way) DropUser() {
	w.UserName = ""
}

func (w *Way) TagsAsMap() map[string]interface{} {
	m := map[string]interface{}{}
	for _, t := range w.Tags {
		m[t.Key] = t.Value
	}
	return m
}
