package osm

import (
	"time"
)

type Element struct {
	Id        int64      `xml:"id,attr"`
	Version   int        `xml:"version,attr,omitempty"`
	Timestamp *time.Time `xml:"timestamp,attr,omitempty"`
	Changeset int64      `xml:"changeset,attr,omitempty"`
	UserId    int64      `xml:"uid,attr,omitempty"`
	UserName  string     `xml:"user,attr,omitempty"`
}

func (e *Element) GetId() int64 {
	return e.Id
}

func (e *Element) DropVersion() {
	e.Version = 0
}

func (e *Element) DropTimestamp() {
	e.Timestamp = nil
}

func (e *Element) DropChangeset() {
	e.Changeset = int64(0)
}

func (e *Element) DropUserId() {
	e.UserId = int64(0)
}

func (e *Element) DropUserName() {
	e.UserName = ""
}
