package osm

import (
	"time"
)

// Element is the base abstract element that Nodes, Ways, and Relations extend.
type Element struct {
	Id        uint64     `xml:"id,attr" parquet:"name=id, type=UINT_64"`                                // The planet-wide unique (literally) id of the element.
	Version   uint16     `xml:"version,attr,omitempty" parquet:"name=version type=UINT_16"`             // The version of this element.  Every modification increments this counter.
	Timestamp *time.Time `xml:"timestamp,attr,omitempty" parquet:"name=timestamp type=TimestampMicros"` // The timestamp of when this element was created or last modified.
	Changeset uint64     `xml:"changeset,attr,omitempty" parquet:"name=changest type=UINT_64"`          // The ID of the changeset that created or modified this element.
	UserId    uint64     `xml:"uid,attr,omitempty" parquet:"name=uid, type=UINT_64"`                    // The id of the user that created or last modified this element.
	UserName  string     `xml:"user,attr,omitempty"`                                                    // The name of the user that created or last modified this element.
}

// GetId returns the element's ID as an int64
func (e *Element) GetId() uint64 {
	return e.Id
}

// DropVersion sets the version to 0
func (e *Element) DropVersion() {
	e.Version = uint16(0)
}

// DropTimestamp sets the timestamp to nil
func (e *Element) DropTimestamp() {
	e.Timestamp = nil
}

// DropChangeset sets the Changeset Id to 0
func (e *Element) DropChangeset() {
	e.Changeset = uint64(0)
}

// DropUserId sets the UserId to 0
func (e *Element) DropUserId() {
	e.UserId = uint64(0)
}

// DropUserName sets the UserName to ""
func (e *Element) DropUserName() {
	e.UserName = ""
}

func (e *Element) DropAttributes(pr *PlanetResource) {
	if pr.DropVersion {
		e.DropVersion()
	}

	if pr.DropTimestamp {
		e.DropTimestamp()
	}

	if pr.DropChangeset {
		e.DropChangeset()
	}

	if pr.DropUserId {
		e.DropUserId()
	}

	if pr.DropUserName {
		e.DropUserName()
	}
}
